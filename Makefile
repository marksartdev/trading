.DEFAULT_GOAL := exchange

.PHONY: up
up:
	docker-compose -f deployments/docker-compose.yaml -p trading up --build -d

.PHONY: down
down:
	docker-compose -f deployments/docker-compose.yaml -p trading down

.PHONY: start
start:
	docker-compose -f deployments/docker-compose.yaml -p trading start

.PHONY: stop
stop:
	docker-compose -f deployments/docker-compose.yaml -p trading stop

.PHONY: exchange
exchange:
	go build -o bin/exchange ./cmd/exchange
	./bin/exchange

.PHONY: broker
broker:
	go build -o bin/broker ./cmd/broker
	./bin/broker
