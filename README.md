# URL Shortener

Сервис для сокращения ссылок с API на Go и отдельным консюмером аналитики. API выдает короткие ссылки и редиректит пользователей, а события кликов публикуются в Kafka и агрегируются консюмером в PostgreSQL.

## Архитектура
- **API** (`cmd/api`) — HTTP‑сервер, который создает короткие ссылки, отдает редиректы и пишет события кликов в Kafka. Кэширует ссылки в Redis.
- **Консюмер аналитики** (`cmd/analytics-consumer`) — читает события кликов из Kafka и сохраняет статистику в PostgreSQL, отдает метрики Prometheus на `:9091/metrics`.
- **Хранилища и инфраструктура**: PostgreSQL для ссылок и аналитики, Redis для кэша ссылок, Kafka для событий, Prometheus и Grafana для наблюдаемости.

## Требования
- Go 1.24+
- Docker и Docker Compose (для запуска всего стека)
- CLI `migrate` для управления миграциями (используется в `Makefile`)

## Переменные окружения
| Переменная | Назначение | Значение по умолчанию |
|------------|------------|-----------------------|
| `APP_ENV` | среда (`dev`, `prod`) | `dev` |
| `HTTP_ADDR` | адрес HTTP‑сервера API | `:8080` |
| `POSTGRES_DSN` | строка подключения к PostgreSQL | — (обязательна) |
| `REDIS_ADDR` | адрес Redis | `localhost:6379` |
| `REDIS_DB` | номер базы Redis | `0` |
| `REDIS_PASSWORD` | пароль Redis | пусто |
| `KAFKA_BROKERS` | список брокеров Kafka через запятую | `localhost:9092` |
| `KAFKA_CLICKS_TOPIC` | имя топика для кликов | `clicks` |
| `BASE_URL` | базовый URL для генерации коротких ссылок | `http://localhost:8080` |
| `KAFKA_CLICKS_CONSUMER_GROUP` | группа консюмера аналитики | `clicks-analytics-consumer` |
| `METRICS_ADDR` | адрес метрик консюмера | `:9091` |

В репозитории лежит готовый `.env.docker` с дефолтами для Docker Compose. Используйте его как есть или измените `BASE_URL` и
другие значения при необходимости. Для локального запуска без Docker можно создать `.env` по аналогии.

## Быстрый старт через Docker Compose

## Quick start

```bash
git clone https://github.com/PavelKhromykhGo/url-shortener
cd url-shortener
docker-compose -f deploy/docker-compose.yml up -d --build
```
   Compose поднимет PostgreSQL, Redis, Kafka, применит миграции, запустит API, консюмер аналитики и мониторинг (Prometheus и Grafana).

После старта API доступен на `http://localhost:8080`, метрики консюмера — на `http://localhost:9091/metrics`, Grafana — на `http://localhost:3000` (логин/пароль `admin/admin`).

## Локальный запуск без Docker
1. Установите `POSTGRES_DSN` и другие переменные окружения.
2. Примените миграции (требуется утилита `migrate`):
   ```bash
   make migrate-up
   ```
3. Запустите API:
   ```bash
   go run ./cmd/api
   ```
4. В отдельном терминале запустите консюмер аналитики:
   ```bash
   go run ./cmd/analytics-consumer
   ```

## Миграции
Миграции расположены в каталоге `migrations` и управляются через `make`:
- `make migrate-up` — применить миграции
- `make migrate-down` — откатить
- `make migrate-drop` — удалить базу данных

## Метрики
- API и консюмер инициализируют метрики Prometheus (`/metrics` для консюмера, Prometheus в compose конфигурируется в `deploy/prometheus/prometheus.yml`).
- Сервис логирует ключевые события через Zap и собственный обертку `internal/logger`.

## Дополнительно
- Генерация коротких кодов происходит через `internal/id` с произвольной длиной (по умолчанию 8 символов).
- Кэш ссылок в Redis подключается автоматически, если Redis доступен; при недоступности сервис продолжит работу без кэша.