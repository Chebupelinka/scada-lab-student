#!/usr/bin/env bash
# Проверка стенда. Лучше запускать с PLC_SPEED>=10, иначе авария наступит через ~1.5 мин.
set -euo pipefail

cd "$(dirname "$0")/.."
PSQL=(docker compose exec -T db psql -U "${POSTGRES_USER:-scada}" -d "${POSTGRES_DB:-scada}" -tA -c)
ok()   { echo "  [OK]   $*"; }
fail() { echo "  [FAIL] $*"; exit 1; }

echo "1. Контейнеры"
for s in plc db gateway fuxa; do
    state=$(docker compose ps --format '{{.State}}' "$s" 2>/dev/null || true)
    [[ "$state" == "running" ]] && ok "$s запущен" || fail "$s не запущен (state='$state'), см. docker compose logs $s"
done

echo "2. Схема БД"
n=$("${PSQL[@]}" "SELECT count(*) FROM tags" 2>/dev/null || echo 0)
[[ "$n" -ge 4 ]] && ok "тегов в справочнике: $n" || fail "в tags меньше 4 тегов (TODO 7)"
n=$("${PSQL[@]}" "SELECT count(*) FROM alarm_limits" 2>/dev/null || echo 0)
[[ "$n" -ge 2 ]] && ok "уставок аварий: $n" || fail "в alarm_limits меньше 2 строк (TODO 5, 7)"
"${PSQL[@]}" "SELECT 1 FROM current_values LIMIT 1" >/dev/null 2>&1 && ok "представление current_values есть" || fail "нет представления current_values (TODO 6)"
n=$("${PSQL[@]}" "SELECT count(*) FROM pg_trigger WHERE tgrelid = 'events'::regclass AND NOT tgisinternal" 2>/dev/null || echo 0)
[[ "$n" -ge 1 ]] && ok "триггер на events есть" || fail "нет триггера NOTIFY на events (TODO 6)"

echo "3. Поступление данных"
for i in $(seq 1 30); do
    n=$("${PSQL[@]}" "SELECT count(*) FROM measurements WHERE time > now() - interval '10 seconds'")
    [[ "$n" -gt 0 ]] && break
    sleep 1
done
[[ "$n" -gt 0 ]] && ok "измерений за 10 с: $n" || fail "шлюз не пишет измерения (TODO 8, 9, 11; docker compose logs gateway)"
t=$("${PSQL[@]}" "SELECT round(value::numeric,1) FROM current_values WHERE name='reactor_temp'")
ok "текущая температура: $t °C"

echo "4. Аварийная сигнализация (ждём превышения H=80, до 180 с)"
for i in $(seq 1 180); do
    n=$("${PSQL[@]}" "SELECT count(*) FROM events WHERE state IN ('HI','HI_HI')")
    [[ "$n" -gt 0 ]] && break
    sleep 1
done
[[ "$n" -gt 0 ]] && ok "событий HI/HI_HI: $n" || fail "событие превышения не сформировано (TODO 2, 10, 11)"
"${PSQL[@]}" "SELECT to_char(time,'HH24:MI:SS') || '  ' || state || '  ' || message FROM events ORDER BY id" | sed 's/^/         /'

echo "5. FUXA"
code=$(curl -s -o /dev/null -w '%{http_code}' "http://localhost:${FUXA_PORT:-1881}/" || true)
[[ "$code" == "200" ]] && ok "веб-интерфейс отвечает (HTTP 200)" || fail "FUXA не отвечает (HTTP $code)"

echo "Все проверки пройдены."
