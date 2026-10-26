package main

import (
	"context"
	"log"
	"net/http"

	"github.com/cuba3/chronos/core/internal/ai"
	"github.com/cuba3/chronos/core/internal/api"
	"github.com/cuba3/chronos/core/internal/config"
	"github.com/cuba3/chronos/core/internal/parser"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// 1. Khởi tạo Parser Client (Rust)
	parserClient, err := parser.NewClient(cfg.ParserAddress)
	if err != nil {
		log.Fatalf("Could not connect to parser: %v", err)
	}
	defer parserClient.Close()

	// 2. Khởi tạo AI Client (Python)
	aiClient, err := ai.NewClient(cfg.AIAddress)
	if err != nil {
		log.Fatalf("Could not connect to AI: %v", err)
	}
	defer aiClient.Close()

	// 3. Setup Router
	r := api.SetupRouter()

	// Endpoint test gọi chéo sang Rust
	r.GET("/api/v1/test-parser", func(c *gin.Context) {
		sourceCode := "package main\nfunc Hello() {}"
		ast, err := parserClient.ParseFile(context.Background(), "test.go", sourceCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":    "parsed",
			"functions": len(ast.Functions),
		})
	})

	// Endpoint test gọi chéo sang Python
	r.GET("/api/v1/test-ai", func(c *gin.Context) {
		resp, err := aiClient.Query(context.Background(), "What does GetUser do?", 3)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"answer":  resp.Answer,
			"sources": len(resp.Sources),
		})
	})

	log.Printf("Go Core API listening on port %s...", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}