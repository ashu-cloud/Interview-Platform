package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/embedding"
)

func main() {
	ctx := context.Background()
	connStr := "postgres://postgres:postgres@localhost:5432/interview_platform?sslmode=disable"
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer pool.Close()

	// 1. Get all blueprints
	rows, err := pool.Query(ctx, `SELECT id, org_id, title FROM assessment_blueprints`)
	if err != nil {
		log.Fatalf("query blueprints failed: %v", err)
	}
	defer rows.Close()

	type BP struct {
		ID    uuid.UUID
		OrgID uuid.UUID
		Title string
	}
	var blueprints []BP
	for rows.Next() {
		var b BP
		if err := rows.Scan(&b.ID, &b.OrgID, &b.Title); err != nil {
			log.Fatalf("scan blueprint failed: %v", err)
		}
		blueprints = append(blueprints, b)
	}
	rows.Close()

	fmt.Printf("Found %d blueprints to index and seed with RAG knowledge chunks\n", len(blueprints))

	for _, bp := range blueprints {
		// Clear existing chunks for clean deterministic reload
		_, _ = pool.Exec(ctx, `DELETE FROM document_chunks WHERE blueprint_id = $1`, bp.ID)

		var chunkTexts []struct {
			Text   string
			Source string
		}

		// 1. Fetch Prompts
		pRows, err := pool.Query(ctx, `SELECT question_text FROM blueprint_prompts WHERE docket_id = $1 ORDER BY order_index`, bp.ID)
		if err == nil {
			for pRows.Next() {
				var q string
				if pRows.Scan(&q) == nil && strings.TrimSpace(q) != "" {
					chunkTexts = append(chunkTexts, struct {
						Text   string
						Source string
					}{Text: fmt.Sprintf("Blueprint Inquiry Focus: %s", q), Source: "blueprint_prompt"})
				}
			}
			pRows.Close()
		}

		// 2. Fetch Criteria
		cRows, err := pool.Query(ctx, `SELECT id, name, description FROM evaluation_criteria WHERE docket_id = $1 ORDER BY order_index`, bp.ID)
		if err == nil {
			for cRows.Next() {
				var cID uuid.UUID
				var name, desc string
				if cRows.Scan(&cID, &name, &desc) == nil {
					text := fmt.Sprintf("Assessment Competency [%s]: %s", name, desc)
					chunkTexts = append(chunkTexts, struct {
						Text   string
						Source string
					}{Text: text, Source: "evaluation_criterion"})

					// Update criterion embedding
					vec := embedding.DeterministicSemanticEmbedding(text, 768)
					_, _ = pool.Exec(ctx, `UPDATE evaluation_criteria SET embedding = $1 WHERE id = $2`, pgvector.NewVector(vec), cID)
				}
			}
			cRows.Close()
		}

		// 3. Add domain architectural reference chunks if relevant to candidate project inquiry (Kafka, Redis, PubSub)
		if strings.Contains(strings.ToLower(bp.Title), "candidate project") || strings.Contains(strings.ToLower(bp.Title), "sde") || strings.Contains(strings.ToLower(bp.Title), "interview") {
			chunkTexts = append(chunkTexts,
				struct {
					Text   string
					Source string
				}{
					Text:   "Kafka vs Redis PubSub Architecture: Apache Kafka is a distributed, horizontally partitioned commit log that persists messages to disk with configurable retention (TTL or compacted), allowing consumer groups to independently track offsets, seek backwards, and replay events. Redis PubSub is an ephemeral in-memory fire-and-forget broadcast where messages are dropped if subscribers are disconnected. Use Kafka for durable event streaming, partition-key ordering, and replayability; use Redis PubSub for transient notifications or Redis Streams for lightweight consumer group queuing.",
					Source: "architecture_knowledge_base",
				},
				struct {
					Text   string
					Source string
				}{
					Text:   "Partition Ordering & Load-Balancing: Kafka guarantees strict message ordering only within a single partition, not globally across the entire topic. To avoid hot partitions when hashing keys (consistent hashing), high-cardinality composite keys or compound partitioners are employed. When consumer lag spikes, rebalancing protocols (CooperativeStickyAssignor) prevent full-stop partitions pauses.",
					Source: "architecture_knowledge_base",
				},
				struct {
					Text   string
					Source string
				}{
					Text:   "Candidate Project Architecture Grounding: When evaluating a candidate's projects, verify their claims regarding concurrency, database isolation levels (Postgres READ COMMITTED vs REPEATABLE READ), handling of duplicate messages via idempotency keys, and failure recovery under network partitions.",
					Source: "evaluation_guidelines",
				},
			)
		}

		// Insert into document_chunks
		inserted := 0
		for i, item := range chunkTexts {
			vec := embedding.DeterministicSemanticEmbedding(item.Text, 768)
			_, err := pool.Exec(ctx, `
				INSERT INTO document_chunks (blueprint_id, org_id, chunk_text, embedding, chunk_index, source)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, bp.ID, bp.OrgID, item.Text, pgvector.NewVector(vec), i, item.Source)
			if err == nil {
				inserted++
			}
		}

		fmt.Printf("Blueprint [%s] (ID %s): inserted %d knowledge chunks\n", bp.Title, bp.ID, inserted)
	}

	// Verify total chunks
	var total int
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM document_chunks`).Scan(&total)
	fmt.Printf("\nSUCCESS: Total document_chunks in PostgreSQL RAG ledger: %d\n", total)
}
