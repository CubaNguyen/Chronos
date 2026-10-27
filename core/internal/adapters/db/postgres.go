package db

import (
	"context"
	"fmt"

	"github.com/cuba3/chronos/core/internal/port"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type PostgresRepo struct {
	pool *pgxpool.Pool
}

// Hàm này trả về Interface, tự động "kế thừa" ngầm định trong Go
func NewPostgresRepo(pool *pgxpool.Pool) port.VectorRepository {
	return &PostgresRepo{pool: pool}
}

func (r *PostgresRepo) SaveEmbeddings(ctx context.Context, items []port.CodeEmbedding) error {
	batch := &pgx.Batch{}
	query := `
		INSERT INTO code_embeddings (
			repo_name, file_path, language, function_name, content, embedding, commit_hash, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (repo_name, file_path, function_name, commit_hash) DO NOTHING;
	`

	for _, item := range items {
		vec := pgvector.NewVector(item.Embedding)
		batch.Queue(query, item.RepoName, item.FilePath, item.Language, item.FunctionName, item.Content, vec, item.CommitHash, item.Metadata)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(items); i++ {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("lỗi khi insert batch record %d: %w", i, err)
		}
	}
	return nil
}

func (r *PostgresRepo) SearchSimilar(ctx context.Context, queryEmbedding []float32, limit int) ([]port.CodeEmbedding, error) {
	query := `
		SELECT id, repo_name, file_path, language, function_name, content, commit_hash
		FROM code_embeddings
		ORDER BY embedding <=> $1
		LIMIT $2
	`
	vec := pgvector.NewVector(queryEmbedding)
	rows, err := r.pool.Query(ctx, query, vec, limit)
	if err != nil {
		return nil, fmt.Errorf("lỗi truy vấn vector: %w", err)
	}
	defer rows.Close()

	var results []port.CodeEmbedding
	for rows.Next() {
		var item port.CodeEmbedding
		err := rows.Scan(&item.ID, &item.RepoName, &item.FilePath, &item.Language, &item.FunctionName, &item.Content, &item.CommitHash)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, nil
}
