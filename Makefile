.PHONY: up down migrate logs build test-jcs test-orchestrator test-platform test-runtime test-evaluation minio-console

up:
	docker compose up -d

down:
	docker compose down

migrate:
	docker compose run --rm migrate

logs:
	docker compose logs -f

build:
	docker compose build

test-jcs:
	cd services/job-candidate-service && go test ./...

test-orchestrator:
	cd services/interview-orchestrator && go test ./...

test-platform: test-jcs

test-runtime: test-orchestrator

test-evaluation:
	cd services/feedback-worker && go test ./...

minio-console:
	@echo "MinIO console: http://localhost:9001 (minioadmin / minioadmin)"
