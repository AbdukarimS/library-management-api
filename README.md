# Library Management API (Go + PostgreSQL + JWT)

Серверное REST API приложение для управления библиотечным фондом, читателями и выдачами книг с постоянным хранением в **PostgreSQL**, поддержкой **SQL-транзакций** и **JWT-авторизацией** на основе ролей (`ADMIN`, `USER`).

## Архитектура проекта
- `cmd/server/main.go` — точка входа, загрузка переменных окружения, инициализация зависимостей и запуск HTTP-сервера.
- `internal/database/postgres.go` — подключение к PostgreSQL через `database/sql` и драйвер `pgx/v5/stdlib`.
- `internal/models/` — структуры данных (`Book`, `Reader`, `Loan`, `User`, `Statistics`).
- `internal/repository/` — слой доступа к данным (параметризованные SQL-запросы и транзакции `BEGIN / COMMIT / ROLLBACK`).
- `internal/services/` — бизнес-логика, валидация, хеширование паролей (`bcrypt`) и генерация JWT.
- `internal/handlers/` — HTTP-обработчики, парсинг запросов и единый JSON-формат ответов.
- `internal/middleware/` — логирование запросов, проверка JWT (`401`) и проверка ролей (`403`).
- `internal/router/router.go` — маршрутизация на стандартном `http.ServeMux` (24 endpoints).
- `migrations/001_init.sql` — SQL-скрипт создания таблиц с ограничениями `PRIMARY KEY`, `FOREIGN KEY`, `NOT NULL`, `UNIQUE`, `CHECK`.

## Настройка и запуск

1. **Настройка PostgreSQL:**
   Создайте базу данных `library_db` и выполните миграцию:
   ```bash
   psql -U postgres -p 5433 -d library_db -f migrations/001_init.sql
   ```
2. **Конфигурация окружения:**
   Скопируйте `.env.example` в `.env` и укажите реквизиты подключения к PostgreSQL и `JWT_SECRET`.
3. **Запуск сервера:**
   ```bash
   go run ./cmd/server
   ```
4. **Проверка сборки и тестов:**
   ```bash
   go build ./...
   go vet ./...
   go test -v ./...
   ```

## Учётная запись администратора и получение JWT
При запуске приложения автоматически создаётся администратор:
- **Email:** `admin@library.local`
- **Password:** `admin123`

Получение токена (`POST /api/auth/login`):
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@library.local","password":"admin123"}'
```
Полученный токен передаётся в заголовке: `Authorization: Bearer <TOKEN>`.

## Таблица API Endpoints (24 маршрута)

| Метод | Endpoint | Доступ | Описание |
|---|---|---|---|
| GET | `/api/health` | Публичный | Проверка состояния сервера |
| GET | `/api/health/db` | Публичный | Проверка подключения к PostgreSQL |
| POST | `/api/auth/register` | Публичный | Регистрация пользователя (`USER`) |
| POST | `/api/auth/login` | Публичный | Аутентификация и выдача JWT |
| GET | `/api/auth/me` | `USER`, `ADMIN` | Профиль текущего пользователя |
| GET | `/api/books` | `USER`, `ADMIN` | Список всех книг |
| GET | `/api/books/{id}` | `USER`, `ADMIN` | Получение книги по ID |
| GET | `/api/books/search` | `USER`, `ADMIN` | Поиск книг по параметрам |
| POST | `/api/books` | `ADMIN` | Добавление книги |
| PUT | `/api/books/{id}` | `ADMIN` | Обновление книги |
| DELETE | `/api/books/{id}` | `ADMIN` | Удаление книги |
| GET | `/api/readers` | `ADMIN` | Список читателей |
| GET | `/api/readers/search` | `ADMIN` | Поиск читателей по имени |
| GET | `/api/readers/{id}` | `ADMIN` | Получение читателя по ID |
| POST | `/api/readers` | `ADMIN` | Регистрация читателя |
| PUT | `/api/readers/{id}` | `ADMIN` | Обновление данных читателя |
| DELETE | `/api/readers/{id}` | `ADMIN` | Удаление читателя |
| GET | `/api/readers/{id}/loans` | `ADMIN` | История выдач читателя |
| GET | `/api/loans` | `ADMIN` | Все операции выдачи |
| GET | `/api/loans/active` | `ADMIN` | Активные выдачи |
| GET | `/api/loans/{id}` | `ADMIN` | Получение выдачи по ID |
| POST | `/api/loans` | `ADMIN` | Выдача книги (транзакция) |
| PUT | `/api/loans/{id}/return` | `ADMIN` | Возврат книги (транзакция) |
| GET | `/api/statistics` | `ADMIN` | Статистика библиотеки |
