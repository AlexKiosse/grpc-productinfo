# ProductInfo — gRPC-микросервис e-commerce

Учебный/демонстрационный проект на **Go**: каталог товаров и управление заказами через **gRPC**, с **mutual TLS (mTLS)**, цепочкой **interceptors**, **PostgreSQL** и **Docker**.

Подходит как референс по связке: Protocol Buffers → gRPC-сервер → клиенты → БД (sqlc + goose).

---

## Возможности

| Компонент | Описание |
|-----------|----------|
| **ProductInfo** | `AddProduct`, `GetProduct` — товары хранятся в PostgreSQL |
| **OrderManagement** | CRUD и стриминг заказов (in-memory, для демо gRPC-паттернов) |
| **Безопасность** | mTLS: сервер и клиент предъявляют сертификаты, подписанные своим CA |
| **Interceptors** | recovery → auth (Bearer + metadata) → logging |
| **Инфраструктура** | Docker Compose: Postgres + сервер + demo-клиенты |

---

## Архитектура

```text
┌─────────────────┐     mTLS gRPC      ┌──────────────────────────────┐
│  client/product │ ─────────────────► │  service (port 50051)        │
│  client/order   │                    │  ├─ ProductInfo  → PostgreSQL │
└─────────────────┘                    │  └─ OrderManagement (memory) │
                                       └──────────────┬───────────────┘
                                                      │
                                                      ▼
                                            ┌─────────────────┐
                                            │  PostgreSQL 16  │
                                            │  table: products│
                                            └─────────────────┘
```

**Модули репозитория**

```text
productinfo/
├── service/          # gRPC-сервер, protobuf, interceptors
├── client/           # CLI-клиенты (product, order)
├── db/
│   ├── migrations/   # goose
│   ├── query/        # SQL для sqlc
│   └── sqlc/         # сгенерированный Go-код
├── certs/            # локальные сертификаты (не в git)
├── docker-compose.yml
└── sqlc.yaml
```

---

## Требования

- **Go** 1.26+
- **Docker** и **Docker Compose**
- **goose** — миграции ([pressly/goose](https://github.com/pressly/goose))
- **sqlc** — генерация кода из SQL (опционально, если меняете запросы)

---

## Быстрый старт (Docker)

Из корня репозитория:

```bash
docker compose up --build
```

Поднимутся Postgres, gRPC-сервер и клиенты (product / order). Сервер ждёт готовности БД и переменную `DATABASE_URL`.

**Миграции** (если Postgres уже запущен отдельно):

```powershell
$env:DATABASE_URL="postgres://admin:admin@localhost:5432/productinfo?sslmode=disable"
goose -dir db/migrations postgres $env:DATABASE_URL up
```

---

## Локальная разработка

### 1. База данных

Только Postgres:

```bash
docker compose -f docker-compose.db.yml up -d
```

Или Postgres из основного `docker-compose.yml`.

### 2. Миграции

```powershell
cd C:\ASC\Go\productinfo
$env:DATABASE_URL="postgres://admin:admin@localhost:5432/productinfo?sslmode=disable"
goose -dir db/migrations postgres $env:DATABASE_URL up
goose -dir db/migrations postgres $env:DATABASE_URL status
```

### 3. Сертификаты

Каталог `certs/` в `.gitignore`. Сертификаты нужно сгенерировать локально (см. историю коммитов про mTLS / скрипты в репозитории, если добавлены).

Для mTLS клиент использует **`client.crt` / `client.key`**, сервер — **`server.crt` / `server.key`**, доверие через **`ca.crt`**.

### 4. Запуск сервера

```powershell
cd service
$env:DATABASE_URL="postgres://admin:admin@localhost:5432/productinfo?sslmode=disable"
go run .
```

Опционально: `CERT_DIR` — каталог с сертификатами (по умолчанию `../certs` относительно `service/`).

### 5. Клиент товаров

```powershell
cd client\product
go run product_client.go
```

Клиент шлёт metadata (`authorization`, `x-request-id`, `x-client-type`) — без них сработает interceptor авторизации.

---

## API (Protocol Buffers)

Определения: `service/ecommerce/*.proto`

**ProductInfo**

- `addProduct(Product) → ProductID`
- `getProduct(ProductID) → Product`

**OrderManagement** — unary и server-streaming RPC для демонстрации работы с заказами.

---

## Стек

- [gRPC](https://grpc.io/) / [protobuf](https://protobuf.dev/)
- [pgx](https://github.com/jackc/pgx) — пул соединений PostgreSQL
- [sqlc](https://sqlc.dev/) — type-safe SQL
- [goose](https://github.com/pressly/goose) — миграции
- [go-grpc-middleware](https://github.com/grpc-ecosystem/go-grpc-middleware) — цепочка interceptors

---

## Переменные окружения

| Переменная | Где | Назначение |
|------------|-----|------------|
| `DATABASE_URL` | service | строка подключения PostgreSQL (обязательна) |
| `CERT_DIR` | service | путь к TLS-сертификатам (по умолчанию `../certs`) |

---

## Полезные команды

```powershell
# Перегенерация sqlc после изменения db/query/*.sql
sqlc generate

# Сборка сервера
cd service
go build -o server.exe .

# Статус миграций
goose -dir db/migrations postgres $env:DATABASE_URL status
```

---

## Безопасность и секреты

- Не коммитьте `*.key` и содержимое `certs/` — в репозитории только `.gitignore`.
- Токен в `client/common/metadata.go` — **демо**, не для production.
- Пароли Postgres в compose — для локальной разработки.
