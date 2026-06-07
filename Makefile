include .env
export

export PROJECT_ROOT=$(CURDIR)

env-up:
	docker compose up -d todoapp-postgres

env-down:
	docker compose down todoapp-postgres

env-cleanup:
	@docker compose down todoapp-postgres; \
	rm -rf out/pgdata; \
	echo "Files were cleaned!";

migrate-create:
	@if [ -z "$(seq)"]; then \
		echo "Add param seq. Example: make migrate-cleaned seq=init"; \
		exit 1; \
	fi; \

	docker compose run --rm	todoapp-postgres-migrate \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

test-target:
	@echo "value: $(var)"