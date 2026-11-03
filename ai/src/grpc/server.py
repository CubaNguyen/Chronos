"""Implement AIService gRPC servicer."""

import logging
from concurrent import futures

import grpc
from grpc_reflection.v1alpha import reflection

from src.config import ensure_dirs
from src.embedder import Embedder
from src.retriever import Retriever
from src.rag import RAG, NO_CONTEXT_MESSAGE
from src.llm import LiteLLMClient
from src.store import VectorStore

from src.grpc import ai_pb2, ai_pb2_grpc
from src.grpc.convert import query_result_to_proto, rag_answer_to_proto

logger = logging.getLogger(__name__)


class AIServiceServicer(ai_pb2_grpc.AIServiceServicer):
    def __init__(self) -> None:
        ensure_dirs()
        logger.info("Loading models...")
        self.embedder = Embedder()
        self.store = VectorStore()
        self.retriever = Retriever(self.embedder, self.store)
        self.llm = LiteLLMClient()
        self.rag = RAG(self.retriever, self.llm)
        logger.info("AIService ready.")

    def Embed(self, request: ai_pb2.EmbedRequest, context: grpc.ServicerContext) -> ai_pb2.EmbedResponse:
        if not request.texts:
            return ai_pb2.EmbedResponse(error="texts is empty")
        try:
            vectors = self.embedder.embed_batch(list(request.texts))
            proto_vectors = [ai_pb2.Vector(values=v) for v in vectors]
            return ai_pb2.EmbedResponse(vectors=proto_vectors)
        except Exception as e:
            logger.exception("Embed failed")
            return ai_pb2.EmbedResponse(error=str(e))

    def Retrieve(self, request: ai_pb2.RetrieveRequest, context: grpc.ServicerContext) -> ai_pb2.RetrieveResponse:
        if not request.query.strip():
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, "query is empty")
        top_k = request.top_k if request.top_k > 0 else 5
        try:
            results = self.retriever.retrieve(request.query, top_k=top_k)
            proto_results = [query_result_to_proto(r) for r in results]
            return ai_pb2.RetrieveResponse(results=proto_results)
        except Exception as e:
            logger.exception("Retrieve failed")
            return ai_pb2.RetrieveResponse(error=str(e))

    def Query(self, request: ai_pb2.QueryRequest, context: grpc.ServicerContext):
        if not request.question.strip():
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, "question is empty")

        # 🚀 Log to see if Go is sending contexts
        logger.info("Received Query. Contexts from Go: %d", len(request.contexts))

        top_k = request.top_k if request.top_k > 0 else 5
        try:
            contexts = None
            if request.contexts:
                from src.models import Chunk, QueryResult
                contexts = [
                    QueryResult(
                        chunk=Chunk(
                            id=c.id,
                            text=c.text,
                            source_type=c.source_type,
                            source_path=c.source_path,
                            metadata=dict(c.metadata),
                        ),
                        score=1.0,
                    )
                    for c in request.contexts
                ]

            # 🚀 ITERATE OVER THE STREAM AND YIELD
            for chunk_text in self.rag.query_stream(request.question, top_k=top_k, contexts=contexts):
                if chunk_text:
                    yield ai_pb2.QueryResponse(
                        answer=chunk_text, 
                        query=request.question
                    )
                    
        except Exception as e:
            logger.exception("RAG query failed")
            yield ai_pb2.QueryResponse(error=str(e), query=request.question)


def serve(port: int = 50052) -> None:
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    ai_pb2_grpc.add_AIServiceServicer_to_server(AIServiceServicer(), server)
    
    service_names = (
        ai_pb2.DESCRIPTOR.services_by_name['AIService'].full_name,
        reflection.SERVICE_NAME,
    )
    reflection.enable_server_reflection(service_names, server)

    server.add_insecure_port(f"[::]:{port}")
    server.start()
    logger.info(f"AI gRPC server listening on port {port}")
    
    try:
        server.wait_for_termination()
    except KeyboardInterrupt:
        logger.info("Shutting down AI server...")
        server.stop(grace=5)