# scada-lab — заготовка

Учебный стенд к практической работе «Развёртывание SCADA и эмуляция ПЛК».
Задание и порядок выполнения — в [методических указаниях](docs/Методические%20указания.pdf).

```
plc (C++, libmodbus) --Modbus TCP :5020--> gateway (Go) --> PostgreSQL
                     \--Modbus TCP-------> FUXA (мнемосхема, аварии)
```

## Требования

- Docker Engine 24+ и Docker Compose v2
- Для разработки без Docker: g++ (C++17), CMake 3.16+, libmodbus-dev, Go 1.26+
- Для проверки: mbpoll, psql (`sudo apt install mbpoll postgresql-client`)

## Что нужно сделать

Готовыми даны: разбор параметров эмулятора, таблица `tags`, типы шлюза (`gateway/types.go`)
и тесты. Остальное пишется по методическим указаниям, места помечены `TODO N`.
Номера идут в порядке выполнения:

| № | Файл | Что написать |
|---|------|-------------|
| 1 | `plc-emulator/src/main.cpp` | `create_mapping`: карта регистров |
| 2 | `plc-emulator/src/main.cpp` | `simulate`: модель реактора |
| 3 | `plc-emulator/src/main.cpp` | `serve`: сервер Modbus TCP на `select()` |
| 4 | `plc-emulator/src/main.cpp` | `main`: запуск и остановка |
| 5 | `db/init/01_schema.sql` | таблицы `measurements`, `alarm_limits`, `events`, `commands` |
| 6 | `db/init/01_schema.sql` | представление `current_values`, триггер NOTIFY |
| 7 | `db/init/02_data.sql` | теги и уставки |
| 8 | `gateway/plc.go` | драйвер ПЛК (simonvetter/modbus) |
| 9 | `gateway/store.go` | хранилище (pgx) |
| 10 | `gateway/alarms.go` | `Evaluate` и `Message` |
| 11 | `gateway/main.go` | цикл опроса |

Проверка:

```bash
cd gateway && go test ./...       # TODO 10 (без стенда интеграционные тесты пропускаются)
cp .env.example .env
docker compose up -d --build
./scripts/check.sh                # весь стенд
cd gateway && MODBUS_URL=tcp://localhost:5020 \
  DATABASE_URL=postgres://scada:scada@localhost:5432/scada go test -v ./...   # TODO 8, 9
```

После изменения `db/init/` базу нужно пересоздать: `docker compose down -v`.

| Сервис  | Адрес                  | Доступ          |
|---------|------------------------|-----------------|
| FUXA    | http://localhost:1881  | без пароля      |
| PostgreSQL | localhost:5432      | scada / scada   |
| ПЛК (Modbus TCP) | localhost:5020 | Unit ID 1      |

## Карта регистров ПЛК

| Тип | Адрес (с 0) | Номер в mbpoll/FUXA (с 1) | Назначение |
|-----|-----|-----|-----|
| Input Register | 0 | 1 (300001) | Температура ×10, int16 |
| Input Register | 1 | 2 (300002) | Мощность нагревателя (факт), % |
| Holding Register | 0 | 1 (400001) | Задание мощности, % (0–100) |
| Holding Register | 1 | 2 (400002) | Амплитуда шума ×10 |
| Coil | 0 | 1 (000001) | Охладитель |
| Coil | 1 | 2 (000002) | Имитация утечки газа |
| Discrete Input | 0 | 1 (100001) | Датчик газа |

Модель: `dT/dt = 0.02·P − 0.008·(T − 20) − 0.03·cooler·(T − 10)`.
`PLC_SPEED` в `.env` ускоряет модельное время.
