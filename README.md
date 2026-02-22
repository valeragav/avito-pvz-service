
[![Go](https://img.shields.io/badge/Go-1.25.5-blue?logo=go)](https://go.dev/)
![CI](https://github.com/valeragav/avito-pvz-service/actions/workflows/ci.yml/badge.svg)
[![codecov](https://codecov.io/github/valeragav/avito-pvz-service/graph/badge.svg?token=P1JNIDX9WR)](https://codecov.io/github/valeragav/avito-pvz-service)

# avito-pvz-service

Решение тестового задания для стажировки в Авито.

[Описание задания](https://github.com/avito-tech/tech-internship/blob/main/Tech%20Internships/Backend/Backend-trainee-assignment-spring-2025/Backend-trainee-assignment-spring-2025.md)

## Сервисы

| Сервис             | URL                            |
| ------------------ | ------------------------------ |
| REST API           | http://localhost:8080          |
| Swagger            | http://localhost:8081/swagger/ |
| Prometheus metrics | http://localhost:9091/metrics  |
| Prometheus UI      | http://localhost:9090/query    |
| Grafana            | http://localhost:3030          |

## Быстрый старт

```bash
docker compose up

make fast-start
```

## Команды

```bash
make help        # список всех команд
make build       # сборка бинарника
make test        # запуск тестов
make coverage    # запуск тестов + отчёт о покрытии
make lint        # запуск линтера
make swagger-init # генерация swagger документации
make bin-deps    # установка зависимостей
```


## Таблицы

```mermaid
erDiagram
    users {
        id UUID PK
        email VARCHAR(255)
        password_hash TEXT
        role VARCHAR(20)
    }

    cities {
        id UUID PK
        name VARCHAR(255)
    }

    pvz {
        id UUID PK
        registration_date TIMESTAMPTZ
        city_id UUID FK
    }

    product_types {
        id UUID PK
        name VARCHAR(255)
    }

    reception_statuses {
        id UUID PK
        name VARCHAR(255)
    }

    receptions {
        id UUID PK
        date_time TIMESTAMPTZ
        pvz_id UUID FK
        status_id UUID FK
    }

    products {
        id UUID PK
        date_time TIMESTAMPTZ
        type_id UUID FK
        reception_id UUID FK
    }

    pvz ||--o{ receptions : "pvz_id"
    cities ||--o{ pvz : "city_id"
    reception_statuses ||--o{ receptions : "status_id"
    receptions ||--o{ products : "reception_id"
    product_types ||--o{ products : "type_id"
```