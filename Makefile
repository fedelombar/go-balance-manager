build:
	@go build -o bin/go-balance-manager

run: build
	@./bin/go-balance-manager

test:
	@go test -v ./...

docker-build:
	@docker build -t go-balance-manager .

docker-run:
	@docker-compose up -d

docker-stop:
	@docker-compose down
