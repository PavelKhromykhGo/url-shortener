include .env
export

MIGRATE := migrate
MIGRATION_PATH := ./migrations
DB_URL := $(POSTGRES_DSN)

migrate-up:
	$(MIGRATE) -path $(MIGRATION_PATH) -database "$(DB_URL)" up

migrate-down:
	$(MIGRATE) -path $(MIGRATION_PATH) -database "$(DB_URL)" down

migrate-drop:
	$(MIGRATE) -path $(MIGRATION_PATH) -database "$(DB_URL)" drop