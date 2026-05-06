# TaskManager API

REST API для управления проектами и задачами.

## Технологии

- Go 1.22+
- SQLite
- Clean Architecture (Handler → Service → Repository)
- Structured Logging (slog)
- Graceful Shutdown

## Запуск

```bash
go run main.go
```

С кастомным конфигом:
```bash
go run main.go config.prod.yaml
```

Через ENV:
```bash
SERVER_PORT=3000 DEBUG=true go run main.go
```

## API

### Users
- `GET /users` — список пользователей
- `POST /users` — создать пользователя

### Projects
- `GET /projects` — список проектов
- `POST /projects` — создать проект

### Tasks
- `GET /projects/{projectId}/tasks?status=todo` — задачи проекта
- `GET /users/{userId}/tasks` — задачи пользователя
- `POST /tasks` — создать задачу
- `PATCH /tasks/{id}` — изменить статус
- `DELETE /tasks/{id}` — удалить задачу

### Health
- `GET /health` — проверка работоспособности

## Конфигурация

```yaml
server:
  host: localhost
  port: 8080
  read_timeout: 15
  write_timeout: 15
  shutdown_timeout: 10

database:
  path: app.db

cors:
  allowed_origins:
    - '*'

debug: true
```

## Тесты

```bash
go test ./...
```