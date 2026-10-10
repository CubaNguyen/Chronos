from src.chunker import Chunker


def test_chunk_count(sample_ast):
    """1 class + 3 functions = 4 chunks."""
    chunks = Chunker().chunk_ast(sample_ast)
    assert len(chunks) == 4


def test_deterministic_id(sample_ast):
    """Cùng input → cùng chunk ID."""
    c1 = Chunker().chunk_ast(sample_ast)
    c2 = Chunker().chunk_ast(sample_ast)
    ids1 = sorted(c.id for c in c1)
    ids2 = sorted(c.id for c in c2)
    assert ids1 == ids2


def test_method_source_type(sample_ast):
    """Method phải có source_type='method', free function là 'function'."""
    chunks = Chunker().chunk_ast(sample_ast)
    by_name = {c.metadata["name"]: c for c in chunks}

    assert by_name["GetUser"].source_type == "method"
    assert by_name["helper"].source_type == "method"
    assert by_name["Add"].source_type == "function"


def test_class_includes_method_signatures(sample_ast):
    """Class chunk phải chứa tên methods của nó."""
    chunks = Chunker().chunk_ast(sample_ast)
    class_chunk = next(c for c in chunks if c.metadata["name"] == "UserService")
    assert "GetUser" in class_chunk.text
    assert "helper" in class_chunk.text
    assert "db *sql.DB" in class_chunk.text