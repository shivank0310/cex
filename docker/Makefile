.PHONY: up down infra build logs ps

COMPOSE_FILE := docker-compose.yml
INFRA_FILE := docker-compose.infra.yml

up:
	docker compose -f $(COMPOSE_FILE) up --build -d

down:
	docker compose -f $(COMPOSE_FILE) down

infra:
	docker compose -f $(INFRA_FILE) up -d

logs:
	docker compose -f $(COMPOSE_FILE) logs -f

ps:
	docker compose -f $(COMPOSE_FILE) ps

build-order:
	docker build -f Dockerfile --build-arg SERVICE_PATH=order-service/cmd/order-service -t cex/order-service ..
