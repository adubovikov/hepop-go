# Конфигурация HEPop-Go

## Общая структура

Конфигурация HEPop-Go хранится в YAML файле и состоит из следующих основных секций:
- Server - настройки HEP сервера
- Writers - настройки систем хранения
- API - настройки HTTP API
- Metrics - настройки Prometheus метрик

## Параметры конфигурации

### Server

- `host` - IP адрес для прослушивания
- `port` - порт для прослушивания
- `protocol` - протокол (udp, tcp, оба)
- `max_packet_size` - максимальный размер пакета
- `read_timeout` - таймаут чтения
- `write_timeout` - таймаут записи
- `workers` - количество рабочих потоков

### Writers

- `type` - тип системы хранения (clickhouse, elastic, loki, multi)
- `batch_size` - размер пакета для записи
- `flush_interval` - интервал сброса буферов

#### ClickHouse

- `host` - IP адрес ClickHouse
- `port` - порт ClickHouse
- `database` - имя базы данных
- `table` - имя таблицы
- `username` - имя пользователя
- `password` - пароль пользователя
- `debug` - включить режим отладки

#### Elasticsearch

- `urls` - список URL Elasticsearch
- `index_name` - имя индекса
- `username` - имя пользователя
- `password` - пароль пользователя
- `debug` - включить режим отладки

#### Loki

- `url` - URL Loki
- `labels` - метки для Loki
- `debug` - включить режим отладки

### API

- `host` - IP адрес для прослушивания
- `port` - порт для прослушивания
- `enable_metrics` - включить Prometheus метрики
- `enable_pprof` - включить pprof профилирование
- `auth_token` - токен для аутентификации
- `cors_origins` - список допустимых CORS источников
- `read_timeout` - таймаут чтения
- `write_timeout` - таймаут записи
