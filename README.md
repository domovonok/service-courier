# service-courier

Go-сервис для управления курьерами и доставками. Сервис хранит данные в PostgreSQL, назначает и освобождает курьеров, отдает HTTP API, собирает метрики Prometheus и взаимодействует с order-сервисом по gRPC. Отдельный worker может обрабатывать события заказов из Kafka.

## Требования

- Go 1.25+
- Docker и Docker Compose
- goose для миграций:

```powershell
go install github.com/pressly/goose/v3/cmd/goose@latest
```

## Подготовка

Создайте `.env` из примера:

```powershell
Copy-Item .env.example .env
```

Для Linux/macOS:

```bash
cp .env.example .env
```

## Запуск через Docker Compose

При значениях из `.env.example`:

```powershell
docker compose up -d postgres
goose -dir ./migrations postgres "postgres://myuser:mypassword@localhost:5432/testdb" up
docker compose up --build service-courier
```

API будет доступен на `http://localhost:8082`.

Для запуска всего состава, включая worker, Prometheus и Grafana:

```powershell
docker compose up --build
```

Для worker нужны доступные Kafka и order-сервис из `.env`: `KAFKA_BROKERS` и `ORDER_SERVICE_HOST`.

## Локальный запуск

```powershell
docker compose up -d postgres
goose -dir ./migrations postgres "postgres://myuser:mypassword@localhost:5432/testdb" up
go run ./cmd/service-courier
```

Локально сервис слушает `http://localhost:8080`.

Worker можно запустить отдельно:

```powershell
go run ./cmd/worker
```

## Полезные команды

```powershell
go test ./...
go build -o service ./cmd/service-courier
docker compose down
```

Makefile:

```powershell
make run
make build
make lint
make down
```

Для `make migrate` и `make run` переменные `GOOSE_DRIVER` и `GOOSE_DBSTRING` должны быть доступны в окружении.

## Адреса

- Docker API: `http://localhost:8082`
- Локальный API: `http://localhost:8080`
- Healthcheck: `HEAD /healthcheck`
- Ping: `GET /ping`
- Metrics: `/metrics`
- Pprof: `http://localhost:6060`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000` (`admin` / `admin`)
