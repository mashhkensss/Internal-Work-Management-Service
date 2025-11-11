# Internal-Work-Management-Service

Internal-Work-Management-Service — это сервис для внутреннего управления заметками, задачами и заказами. Он предоставляет HTTP API с валидацией, журналированием, JWT-аутентификацией и поддержкой идемпотентности при создании заказов. Сервис строится как учебный пример продакшн-ориентированного Go-приложения: конфигурация через переменные окружения, явное построение зависимостей, разнесение домена, сервиса и инфраструктуры, а также готовые миграции PostgreSQL и OpenAPI-спецификация.

## Возможности API

| Ресурс | Методы | Особенности |
| --- | --- | --- |
| `/live` | `GET` | Простой health-check, доступен без авторизации. |
| `/notes` | `POST`, `GET`, `GET /{id}`, `DELETE /{id}` | CRUD для заметок. |
| `/tasks` | `POST`, `GET`, `POST /{id}/done` | Управление задачами. Список поддерживает фильтрацию по `done=true/false`. |
| `/orders` | `POST`, `GET`, `GET /{id}`, `DELETE /{id}` | Работа с заказами и позициями. Создание использует заголовок `Idempotency-Key` для защиты от повторов и стратегию скидок. |

Все конечные точки (кроме `/live`) защищены middleware `Authorization: Bearer <jwt>`, которое ожидает токен в формате HS256 и проверяет подпись с секретом `JWT_SECRET`.

## Архитектура и основные пакеты

- `cmd/app` — точка входа, инициализация логгера и HTTP-сервера.
- `internal/app` — конфигурация, bootstrap, middleware (RequestID, logging, JWT, recovery) и HTTP-обработчики на базе `chi`.
- `internal/domain` — бизнес-модели заметок, задач, заказов и скидок.
- `internal/service` — бизнес-логика, валидация данных, применение скидок и идемпотентности.
- `internal/infrastructure/persistence/postgres` — репозитории на `pgx/v5` и SQL builder `squirrel`.
- `internal/infrastructure/persistence/in_memory` — простые реализации для тестов.
- `internal/factory` — фабрики, связывающие сервисы и репозитории.
- `migrations` — SQL для создания схемы (`notes`, `tasks`, `orders`, `order_items`, `idempotency`).
- `openapi/openapi.yaml` — описание API, совместимое со Swagger UI.
- `tests` и файлы `_tests.go` — набор юнит/интеграционных тестов.

## Технологический стек

- Go 1.21+
- `github.com/go-chi/chi/v5` — HTTP-роутер
- `github.com/jackc/pgx/v5` + `github.com/Masterminds/squirrel` — доступ к PostgreSQL
- `gotest.tools/v3`, in-memory репозитории и заглушки — тестирование
- Docker Compose — локальная БД
- Swagger UI — просмотр OpenAPI-спецификации

## Подготовка окружения

### Зависимости

- Go 1.21 или новее
- Docker + Docker Compose
- `psql` (для применения миграций)

### PostgreSQL

```bash
docker-compose up -d
docker-compose ps    # статус postgres должен быть healthy
```

### Миграции

```bash
psql postgres://app:app@localhost:5432/app -f migrations/01_init.up.sql
```

### Переменные окружения

| Имя | Обязательна | Значение по умолчанию | Назначение |
| --- | --- | --- | --- |
| `HTTP_ADDR` | Нет | `:8080` | Адрес HTTP-сервера. |
| `HTTP_READ_TIMEOUT` | Нет | `5s` | Таймаут чтения запроса. |
| `HTTP_WRITE_TIMEOUT` | Нет | `10s` | Таймаут отправки ответа. |
| `HTTP_SHUTDOWN_TIMEOUT` | Нет | `10s` | Таймаут graceful shutdown. |
| `POSTGRES_DSN` | Да | — | Подключение к БД, например `postgres://app:app@localhost:5432/app?sslmode=disable`. |
| `JWT_SECRET` | Да | — | Секрет для проверки HS256 JWT. |

## Локальный запуск

```bash
export POSTGRES_DSN=postgres://app:app@localhost:5432/app?sslmode=disable
export JWT_SECRET=dev-secret
go run ./cmd/app
```

При успешном старте в логах появится `bootstrap complete` и параметры HTTP-сервера. Для проверки доступности используйте:

```bash
curl http://localhost:8080/live
# ok
```

## Примеры запросов

Во всех примерах ниже используйте валидный JWT-токен с `Authorization: Bearer <token>` (например, создайте его на https://jwt.io с payload `{"sub":"demo","exp":<timestamp>}` и секретом `JWT_SECRET`).

```bash
# Добавить задачу
curl -X POST http://localhost:8080/tasks \
     -H 'Authorization: Bearer <token>' \
     -H 'Content-Type: application/json' \
     -d '{"title":"Prepare report","description":"Finish internal backlog review"}'

# Отметить задачу выполненной
curl -X POST http://localhost:8080/tasks/1/done \
     -H 'Authorization: Bearer <token>'

# Создать заказ c защитой от повторов
curl -X POST http://localhost:8080/orders \
     -H 'Authorization: Bearer <token>' \
     -H 'Content-Type: application/json' \
     -H 'Idempotency-Key: my-key-1' \
     -d '{"customer":"Иван Иванов","items":[{"name":"Ноутбук","price":1000},{"name":"Мышка","price":50}]}'

# Добавить заметку
curl -X POST http://localhost:8080/notes \
     -H 'Authorization: Bearer <token>' \
     -H 'Content-Type: application/json' \
     -d '{"title":"Test note","content":"Hello"}'
```

## Тесты

```bash
go test ./...
# или запустить только HTTP-слой:
go test ./internal/app/handlers/...
```

## Документация API

Спецификация лежит в `openapi/openapi.yaml`. Для просмотра через Swagger UI:

```bash
docker run --rm -p 8081:8080 \
  -e SWAGGER_JSON=/spec/openapi.yaml \
  -v "$(pwd)"/openapi:/spec \
  swaggerapi/swagger-ui
```

UI будет доступен на http://localhost:8081.

## Структура проекта

```
.
├── cmd/app                 # entrypoint приложения
├── internal
│   ├── app                 # config, bootstrap, middleware, HTTP-хендлеры и роутер
│   ├── domain              # модели и бизнес-правила
│   ├── service             # бизнес-логика заметок, задач, заказов
│   ├── infrastructure      # logger, postgres и in-memory репозитории
│   └── factory             # сборка репозиториев и сервисов
├── migrations              # SQL-миграции схемы БД
├── openapi                 # описание публичного API
└── tests                   # общие тестовые хелперы
```
