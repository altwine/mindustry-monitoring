# mindustry-monitoring

# Запуск
**Перед запуском:** создайте файл `.env` с `DISCORD_TOKEN=...` внутри проекта.

## Через `go run` (dev)
```bash
$ go run .
```

## Через docker-контейнер (prod)
```bash
$ docker build -t mindustry-monitoring .
$ docker run --rm --env-file .env -p 8080:8080 mindustry-monitoring
```
