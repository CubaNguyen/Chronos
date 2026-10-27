CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE code_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repo_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(1024) NOT NULL,
    language VARCHAR(50),
    function_name VARCHAR(255),       
    content TEXT NOT NULL,           
    embedding vector(384),            -- Model HuggingFace thường 384 chiều (vd: all-MiniLM-L6-v2)
    commit_hash VARCHAR(40) NOT NULL,
    metadata JSONB,                   
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (repo_name, file_path, function_name, commit_hash)
);

-- Index tối ưu cho Vector Search (HNSW)
CREATE INDEX idx_code_embeddings_hnsw ON code_embeddings 
USING hnsw (embedding vector_cosine_ops) 
WITH (m = 16, ef_construction = 64);

-- Index cho tìm kiếm từ khóa (Full-Text Search)
CREATE INDEX idx_code_embeddings_fts ON code_embeddings 
USING GIN (to_tsvector('english', content));

-- Index filter
CREATE INDEX idx_code_embeddings_repo ON code_embeddings(repo_name);