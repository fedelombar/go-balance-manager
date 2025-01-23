include .env
export $(shell sed 's/=.*//' .env)

build:
	go build -o bin/go-balance-manager

run: build
	./bin/go-balance-manager -addr=$(APP_ADDR) -dbhost=$(DB_HOST) -dbport=$(DB_PORT) -dbuser=$(DB_USER) -dbpass=$(DB_PASS) -dbname=$(DB_NAME)

run-seed: build
	./bin/go-balance-manager -addr=$(APP_ADDR) -dbhost=$(DB_HOST) -dbport=$(DB_PORT) -dbuser=$(DB_USER) -dbpass=$(DB_PASS) -dbname=$(DB_NAME) -seed=true

test:
	go mod tidy
	go test -v ./...

test-integration:
	docker-compose down -v
	docker-compose up -d
	sleep 5 # waiting containers to init
	go test -v -tags=integration ./...

docker-build:
	docker build -t go-balance-manager .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

docker-seed:
	docker-compose run --rm app ./go-balance-manager -seed=true
