# WB Orders Demo

Сервис принимает заказы из Kafka, валидирует их, сохраняет в PostgreSQL и отдаёт через HTTP API. Есть кэш в памяти и вспомогательный эндпоинт для ручной записи заказа без Kafka.

---

## Структура

awesomeProject/
├── app/
│ ├── main.go — HTTP API + Kafka consumer
│ ├── repository.go — доступ к БД (pgxpool)
│ ├── repository_mem.go — репозиторий в памяти
│ ├── validator.go — валидация заказов
│ ├── model.go — модели сущностей
│ ├── cache.go — LRU-кэш заказов
│ └── gen/ — генератор тестовых данных
│ ├── gen.go
│ └── gen_test.go
├── sql/
│ ├── 001_init.sql — миграция up
│ └── 001_init_down.sql — миграция down
├── static/ — статика (html/css)
├── docker-compose.yml
├── .env.example
└── README.md


---

## Технологии

- Go 1.22
- Gorilla Mux (HTTP)
- PostgreSQL + pgx/v5 (pgxpool)
- Kafka (Bitnami) + segmentio/kafka-go
- go-playground/validator/v10 (валидация)
- brianvoe/gofakeit (генерация данных)

---

## Быстрый старт (всё в Docker)

1) Создай `.env`:
```bash
cp .env.example .env
Подними стек:

docker compose up -d postgres zookeeper kafka app
Проверка здоровья:


curl -i http://localhost:8081/healthz
Доступы:

API: http://localhost:8081

Kafka UI: http://localhost:8085

Postgres: localhost:5432 (user: demo, db: orders)

Альтернатива: приложение локально, инфраструктура в Docker
.env для локального запуска приложения:


HTTP_ADDR=:8081
PG_HOST=localhost
PG_PORT=5432
PG_DB=orders
PG_USER=demo
PG_PASSWORD=demo
PG_SSLMODE=disable
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=orders
KAFKA_GROUP=wb-orders-demo
Поднять БД и Kafka:

docker compose up -d postgres zookeeper kafka
Запуск приложения:

cd app
go run .
Миграции
Применить:

psql -h localhost -U demo -d orders -f sql/001_init.sql
Откатить:

psql -h localhost -U demo -d orders -f sql/001_init_down.sql
(В docker-compose миграция 001_init.sql также выполняется автоматически при первом старте Postgres.)

Kafka flow
Producer пишет JSON заказа в топик orders.

Consumer (приложение) читает, валидирует и пишет в PostgreSQL.

API выдаёт заказ по order_uid.

Отправка сообщения в Kafka (Bitnami):


docker compose exec -T kafka bash -lc "/opt/bitnami/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --create --if-not-exists --topic orders --replication-factor 1 --partitions 1"
echo '{"order_uid":"ORD-PIPE", "...": "..."}' | docker compose exec -T kafka bash -lc "/opt/bitnami/kafka/bin/kafka-console-producer.sh --bootstrap-server localhost:9092 --topic orders"
Проверка:

curl http://localhost:8081/api/orders/ORD-PIPE
Эндпоинты
GET  /healthz
POST /api/validate          — проверка валидности заказа (эхо-ответ при успехе)
POST /api/orders            — ручная запись заказа без Kafka
GET  /api/orders/{order_uid}
Пример ручной записи:

curl -X POST http://localhost:8081/api/orders \
  -H "Content-Type: application/json" \
  -d @order.json
Тесты
Генератор и базовые тесты:

cd app
go test ./gen -v
Конфигурация через ENV
Приложение собирает DSN из переменных:

HTTP_ADDR
PG_HOST, PG_PORT, PG_DB, PG_USER, PG_PASSWORD, PG_SSLMODE
KAFKA_BROKERS, KAFKA_TOPIC, KAFKA_GROUP
Режимы:

REPO=inmem         — репозиторий в памяти (без БД)
KAFKA_DISABLED=1   — отключить консьюмер Kafka
Типичный сценарий проверки

# 1) поднять стек
docker compose up -d postgres zookeeper kafka app

# 2) отправить заказ в Kafka
echo '{"order_uid":"ORD-PIPE", "...": "..."}' | docker compose exec -T kafka bash -lc '/opt/bitnami/kafka/bin/kafka-console-producer.sh --bootstrap-server localhost:9092 --topic orders'

# 3) получить заказ из API
curl http://localhost:8081/api/orders/ORD-PIPE

Если нужно — добавлю раздел про архитектурные принципы (слои, интерфейсы, мок-генерация) и примеры unit-тестов для р