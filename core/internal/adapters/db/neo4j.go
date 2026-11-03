package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cuba3/chronos/core/internal/port"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Neo4jRepo struct {
	driver neo4j.DriverWithContext
}

// NewNeo4jConn initializes the connection to Neo4j
func NewNeo4jConn(uri string) (neo4j.DriverWithContext, error) {
	// Because you set NEO4J_AUTH=none in docker-compose, we use neo4j.NoAuth()
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.NoAuth())
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	// Verify the connection actually works
	ctx := context.Background()
	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to Neo4j: %w", err)
	}

	slog.Info("Successfully connected to Neo4j!")
	return driver, nil
}

// NewNeo4jRepo creates the repository adapter
func NewNeo4jRepo(driver neo4j.DriverWithContext) port.GraphRepository {
	return &Neo4jRepo{driver: driver}
}

func (r *Neo4jRepo) SaveAST(ctx context.Context, repoName string, ast *port.Ast) error {
	// 1. Create a session (Like a temporary workspace for this specific task)
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	// 2. Run everything inside a transaction
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {

		// 🚀 STEP A: Create the File Node
		fileQuery := `
			MERGE (f:File {path: $path, repo: $repo})
			SET f.language = $language
		`
		_, err := tx.Run(ctx, fileQuery, map[string]any{
			"path":     ast.File.Path,
			"repo":     repoName,
			"language": ast.File.Language,
		})
		if err != nil {
			return nil, err
		}

		// 🚀 STEP B: Create Function Nodes and link them to the File
		for _, fn := range ast.Functions {
			funcQuery := `
				MATCH (f:File {path: $filePath, repo: $repo})
				MERGE (fn:Function {name: $funcName, repo: $repo})
				SET fn.line_start = $lineStart, fn.line_end = $lineEnd
				MERGE (fn)-[:BELONGS_TO]->(f)
			`
			_, err := tx.Run(ctx, funcQuery, map[string]any{
				"filePath":  ast.File.Path,
				"repo":      repoName,
				"funcName":  fn.Name,
				"lineStart": fn.LineStart,
				"lineEnd":   fn.LineEnd,
			})
			if err != nil {
				return nil, err
			}
		}

		// 🚀 STEP C: Create CALLS relationships
		// If Function A calls Function B, we draw an arrow between them
		for _, call := range ast.Calls {
			callQuery := `
				MERGE (caller:Function {name: $caller, repo: $repo})
				MERGE (callee:Function {name: $callee, repo: $repo})
				MERGE (caller)-[:CALLS]->(callee)
			`
			_, err := tx.Run(ctx, callQuery, map[string]any{
				"caller": call.Caller,
				"callee": call.Callee,
				"repo":   repoName,
			})
			if err != nil {
				return nil, err
			}
		}

		return nil, nil // Transaction successful
	})

	if err != nil {
		slog.Error("Failed to save AST to Neo4j", "file", ast.File.Path, "error", err)
		return fmt.Errorf("neo4j transaction failed: %w", err)
	}

	slog.Info("Successfully saved AST to Neo4j", "file", ast.File.Path)
	return nil
}

func (r *Neo4jRepo) GetCallers(ctx context.Context, functionName string) ([]port.GraphNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	// 🚀 The Magic Cypher Query: Find callers up to 3 levels deep
	query := `
		MATCH (caller:Function)-[:CALLS*1..3]->(target:Function {name: $funcName})
		RETURN DISTINCT caller.name AS name, caller.repo AS repo
		LIMIT 50
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx, query, map[string]any{"funcName": functionName})
		if err != nil {
			return nil, err
		}

		var nodes []port.GraphNode
		for res.Next(ctx) {
			record := res.Record()
			name, _ := record.Get("name")
			repo, _ := record.Get("repo")

			nodes = append(nodes, port.GraphNode{
				Name: name.(string),
				Type: "Function",                          // We know it's a function from the query
				ID:   repo.(string) + ":" + name.(string), // Fake ID for now
			})
		}
		return nodes, res.Err()
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query callers: %w", err)
	}

	// Type assertion because ExecuteRead returns `any`
	nodes, ok := result.([]port.GraphNode)
	if !ok {
		return nil, nil // No callers found
	}
	return nodes, nil
}
