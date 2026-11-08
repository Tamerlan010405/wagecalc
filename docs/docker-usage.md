# Запуск приложения через Docker

Все команды нужно выполнять из корня репозитория `wagecalc-app`, где лежит файл `docker-compose.yml`.

## Запустить приложение

```bash
docker compose up --build -d
```

После запуска:

- фронтенд: http://localhost:3000
- API: http://localhost:8080
- PostgreSQL: localhost:5432

Письма восстановления в локальном Docker-окружении доступны в Mailpit:
http://localhost:8025

При первом запуске Docker скачает нужные образы, соберет фронтенд и бэкенд, поднимет PostgreSQL и выполнит миграции.

## Закрыть приложение, но оставить БД

```bash
docker compose down
```

Эта команда остановит и удалит контейнеры приложения, но данные PostgreSQL останутся в Docker volume `postgres_data`.
Docker обычно отображает этот volume с префиксом проекта: `wagecalc-app_postgres_data`.
При следующем запуске приложение поднимется с прежней базой данных.

Если нужно только остановить контейнеры, не удаляя их:

```bash
docker compose stop
```

## Очистить БД и закрыть приложение

```bash
docker compose down -v
```

Эта команда остановит приложение и удалит Docker volume с данными PostgreSQL.
При следующем запуске база будет создана заново, а миграции выполнятся на чистую БД.

## Перезапустить после остановки

Если приложение было закрыто командой `docker compose down`:

```bash
docker compose up --build -d
```

Если приложение было только остановлено командой `docker compose stop`:

```bash
docker compose start
```


Если надо пересобрать проект после внесения изменений:
```
docker compose down
docker compose build --no-cache
docker compose up -d
```

Чтобы посмотреть логи:

```
docker logs -f wagecalc-api
```

