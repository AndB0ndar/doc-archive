# Go Backend

Бэкенд на Go для системы интеллектуального поиска по PDF‑документам.  
Обеспечивает загрузку файлов, полнотекстовый и семантический поиск, асинхронную обработку документов через очередь задач.

---

## Основные возможности

- Загрузка PDF (до 50 МБ) с метаданными (название, автор, год, категория)
- Сохранение файлов в MinIO (S3‑совместимое хранилище)
- Извлечение текста и разбиение на смысловые чанки
- Генерация эмбеддингов через отдельный Python‑микросервис
- Два режима поиска:
  - **Полнотекстовый** – на основе триграмм (pg_trgm) с ранжированием
  - **Семантический** – через косинусное сходство векторов (pgvector)
- JWT‑аутентификация, разделение документов по пользователям
- Асинхронная обработка через очередь **Asynq** (Redis)
- Кэширование метаданных и сессий в Redis
- Graceful shutdown, структурированное логирование (slog)

---

## Технологический стек

| Компонент          | Библиотека / Инструмент                     |
|--------------------|---------------------------------------------|
| Маршрутизация      | `chi`                                       |
| Работа с БД        | `pgx/v5` + `pgvector-go`                    |
| Миграции           | `golang-migrate/migrate`                    |
| Очередь задач      | `hibiken/asynq` (поверх Redis)              |
| Объектное хранилище| `minio-go/v7`                               |
| Кэш                | `redis/go-redis/v9`                         |
| Парсинг PDF        | `ledongthuc/pdf`                            |
| Логирование        | `log/slog`                                  |
| Конфигурация       | `spf13/viper` (опционально) или `os.Getenv` |

---

## Структура проекта

```
server/
├── cmd/
│   ├── api/                 # точка входа HTTP‑сервера
│   └── worker/              # точка входа Asynq‑воркера
├── internal/
│   ├── transport/           # транспортный слой
│   │   ├── http/            # HTTP‑обработчики, middleware, роутер
│   │   └── queue/           # Asynq‑обработчики и роутер задач
│   ├── service/             # бизнес‑логика
│   ├── repository/          # слой доступа к данным (PostgreSQL)
│   ├── domain/              # бизнес‑сущности и интерфейсы
│   ├── infrastructure/      # адаптеры: cache, queue, storage, clients, parser
│   ├── tasks/               # типы задач для очереди (payload, константы)
│   ├── config/              # загрузка конфигурации
│   └── migrations/          # SQL‑миграции (обычно выносятся в корень)
├── go.mod
├── go.sum
└── README.md
```

---

## Требования к окружению

- Go 1.21+
- PostgreSQL 15+ с расширениями `vector` и `pg_trgm`
- Redis 7+
- MinIO (или любое S3‑совместимое хранилище)
- Python‑сервис эмбеддингов (см. `../embedder`)

> Для разработки все зависимости можно поднять через `docker-compose.infra.yml` из корня проекта.

---

## Переменные окружения

| Переменная               | Описание                                    | Значение по умолчанию                     |
|--------------------------|---------------------------------------------|-------------------------------------------|
| `DATABASE_URL`           | PostgreSQL connection string                | `postgres://user:pass@localhost:5432/docdb?sslmode=disable` |
| `REDIS_URL`              | Адрес Redis (хост:порт)                     | `localhost:6379`                          |
| `MINIO_URL`              | MinIO endpoint                              | `localhost:9000`                          |
| `MINIO_ACCESS_KEY`       | Access Key MinIO                            | `minioadmin`                              |
| `MINIO_SECRET_KEY`       | Secret Key MinIO                            | `minioadmin`                              |
| `MINIO_BUCKET`           | Имя bucket для PDF                          | `pdf-documents`                           |
| `EMBEDDER_URL`           | Адрес Python‑сервиса эмбеддингов            | `http://localhost:5001`                   |
| `PORT`                   | Порт HTTP‑сервера                           | `8080`                                    |
| `ENV`                    | Окружение (development / production)        | `development`                             |
| `SECRET_KEY`             | Секрет для подписи JWT                      | *обязателен*                              |
| `QUEUE_CONCURRENCY`      | Кол-во параллельных воркеров Asynq          | `5`                                       |
| `QUEUE_MAX_RETRY`        | Максимальное число повторных попыток задачи | `3`                                       |

---

## Запуск

### 1. Только Go‑сервер (без Docker)

Убедитесь, что PostgreSQL, Redis, MinIO и Python‑сервис запущены.

```bash
# Установка зависимостей
go mod download

# Применение миграций (создаст таблицы)
go run cmd/api/main.go -migrate   # или используйте отдельную команду migrate

# Запуск HTTP‑сервера
go run cmd/api/main.go

# В другом терминале – запуск воркера
go run cmd/worker/main.go
```

### 2. С Docker (рекомендуется)

Из корня проекта (там где `docker-compose.infra.yml` и `docker-compose.app.yml`):

```bash
# Поднять инфраструктуру (БД, Redis, MinIO, embedder)
docker-compose -f docker-compose.infra.yml up -d

# Поднять Go‑сервисы (api, worker) и веб‑интерфейс
docker-compose -f docker-compose.app.yml up -d
```

---

## Миграции базы данных

Миграции управляются через `golang-migrate`.  
Файлы миграций лежат в `../migrations/` (относительно корня проекта).

Применение миграций вручную:
```bash
migrate -path ../migrations -database "$DATABASE_URL" up
```

При запуске `cmd/api/main.go` с флагом `-migrate` миграции применяются автоматически.

---

## API

Полная документация (OpenAPI) доступна после запуска сервера по адресу:  
`http://localhost:8080/swagger/index.html`

Основные эндпоинты:

| Метод | Путь                               | Описание                            |
|-------|------------------------------------|-------------------------------------|
| POST  | `/register`                        | Регистрация пользователя            |
| POST  | `/login`                           | Вход, получение JWT                 |
| POST  | `/upload`                          | Загрузка PDF                        |
| GET   | `/search`                          | Поиск (параметры `q`, `type`, `limit`) |
| GET   | `/documents`                       | Список документов пользователя      |
| GET   | `/documents/{id}`                  | Метаданные документа                |
| GET   | `/documents/{id}/download-url`     | Временная ссылка на PDF (15 мин)    |
| GET   | `/documents/{id}/download`         | Скачать документ                    |
| DELETE| `/documents/{id}`                  | Удаление документа                  |

Все эндпоинты, кроме `/register` и `/login`, требуют заголовок:  
`Authorization: Bearer <JWT>`

---

## Очередь задач (Asynq)

- **Тип задачи:** `process:document`
- **Payload:** `{"document_id": "...", "object_key": "..."}`
- **Обработчик:** `transport/queue/handlers/document.go` → вызывает `service.DocumentService.ProcessDocument`

Задача автоматически повторяется при ошибке (до `QUEUE_MAX_RETRY` раз). Статус документа обновляется в БД: `pending` → `processing` → `done` / `error`.

---

## Разработка и тестирование

### Форматирование и линтинг
```bash
go fmt ./...
go vet ./...
```

### Тесты
```bash
go test ./... -v
```

Для интеграционных тестов требуется запущенный PostgreSQL (можно использовать `testcontainers-go`).

---

## Сборка и production

```bash
# Сборка бинарников
go build -o bin/api cmd/api/main.go
go build -o bin/worker cmd/worker/main.go
```

Для продакшена рекомендуется использовать готовый `Dockerfile` (предполагается многостадийная сборка).

