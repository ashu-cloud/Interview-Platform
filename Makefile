.PHONY: up down migrate logs build test-jcs test-orchestrator

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
