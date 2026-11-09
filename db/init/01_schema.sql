CREATE TABLE tags (
    id          serial PRIMARY KEY,
    name        text UNIQUE NOT NULL,
    description text NOT NULL,
    unit        text NOT NULL DEFAULT '',
    reg_type    text NOT NULL CHECK (reg_type IN ('coil', 'discrete', 'holding', 'input', 'calc')),
    address     int  NOT NULL CHECK (address >= 0),
    scale       double precision NOT NULL DEFAULT 1.0   -- value = raw * scale
);

-- TODO 5: таблицы measurements (с индексом по tag_id, time), alarm_limits, events, commands.

-- TODO 6: представление current_values, функция notify_event() и триггер на events.
