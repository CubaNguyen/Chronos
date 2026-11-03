package port

import "context"

// GraphNode represents a node in our Knowledge Graph
type GraphNode struct {
	ID   string
	Name string
	Type string // e.g., "Function", "File", "Class"
}

type GraphRepository interface {
	// SaveAST takes the AST from Rust and saves nodes/relationships to Neo4j
	SaveAST(ctx context.Context, repoName string, ast *Ast) error

	// (For F8 - Impact Analysis) Find which functions call this function
	GetCallers(ctx context.Context, functionName string) ([]GraphNode, error)
}
