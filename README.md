<p align="center">
	<img src="./assets/icon.png" alt="Mindustry monitoring icon"/>
</p>

---

# mindustry-monitoring
Сбор статистики публичных серверов mindustry и её отображение через дискорд-бота.

> [!WARNING]
> Не готово для продакшена

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

# Лицензия
MIT, [LICENSE](LICENSE).
