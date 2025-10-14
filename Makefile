.PHONY: help up down seed api executor risk reports bot web dev

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?##' \
		$(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-12s\033[0m %s\n", $$1, $$2}'

up: ## docker compose up Timescale+Redis
	docker compose up -d timescale redis

down: ## compose down -v
	docker compose down -v

seed: ## apply schema to Timescale
	psql "$$TIMESCALE_DSN" -f infra/sql/schema.sql

api: ## run Go supervisor service
	cd apps/supervisor && GOTOOLCHAIN=local go run ./...

executor: ## run Go trade executor service
	cd apps/executor && GOTOOLCHAIN=local go run ./...

risk: ## run Go risk service
	cd apps/risk && GOTOOLCHAIN=local go run ./...

reports: ## run Go reporting service
	cd apps/reports && GOTOOLCHAIN=local go run ./...

bot: ## run Python bot locally
	cd bots/python/samples && python main.py

web: ## run dashboard dev server
	npm --prefix apps/dashboard run dev

dev: ## create dev container
	docker compose up dev-env
