# Go Balance Manager

## Description

This code provides a solution to the problem outlined in the file problem.md

## Features

- User Balance Management
- Idempotent Transactions
- Automated Testing
- Docker Containerization
- Environment Configuration

## Prerequisites

- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [Make](https://www.gnu.org/software/make/) (optional, just more fun and easy)

## Installation and Setup

### 1. Clone the Repository

```bash
git clone https://github.com/your_username/go-balance-manager.git
cd go-balance-manager
```

### 2. Configure Environment Variables

```bash
cp .env.example .env
```

Edit `.env` with your configurations:

```env
Database Configuration

DB_HOST=localhost       # Hostname for the PostgreSQL database
DB_PORT=5433            # Port for the PostgreSQL database
DB_USER=golang_user     # Username for database access
DB_PASS=golang_pass     # Password for database access
DB_NAME=golang_db       # Name of the database

Application Configuration

APP_ADDR=:8080          # Port and address for the application to listen on

Seeding Option

SEED=false              # Set to true to seed predefined users on startup
```

## Running the Application

### Start Application

```bash
docker-compose up -d
```

### Access Application

Dockerized Application: http://localhost:8081

### Seed Database

```bash
make docker-seed
```

## Testing

### Run Unit Tests

```bash
make test
```

### Run Integration Tests

```bash
make test-integration
```

### Run Load Tests

```bash
make load-test
```

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Compiles the Go application |
| `make run` | Builds and runs the application locally |
| `make test` | Runs unit tests |
| `make test-integration` | Runs integration tests |
| `make docker-build` | Builds Docker image |
| `make docker-run` | Starts Docker containers |
| `make load-test` | Executes load tests |

## Troubleshooting

- Check container status: `docker-compose ps`
- View application logs: `docker logs go_balance_manager_app`
- View database logs: `docker logs go_balance_manager_db`

## Security Considerations

- Protect `.env` file
- Use strong database passwords
- Limit database permissions

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

MIT License

## Contact

Federico Lombardozzi
