.PHONY: local-dev

# Start database, run migrations, and start the app
local-dev:
	docker compose up --build
