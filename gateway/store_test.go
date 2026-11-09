package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func testDB(t *testing.T) (*Store, *pgx.Conn) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL не задан")
	}
	ctx := context.Background()
	s, err := NewStore(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close(ctx) })
	return s, conn
}

// testTag создаёт временный тег с уставкой hi и удаляет его после теста.
func testTag(t *testing.T, conn *pgx.Conn, hi float64) string {
	ctx := context.Background()
	name := fmt.Sprintf("test_%d", time.Now().UnixNano())
	var id int
	err := conn.QueryRow(ctx, `INSERT INTO tags (name, description, unit, reg_type, address, scale)
		VALUES ($1, 'Тестовый тег', 'ед', 'holding', 1, 0.5) RETURNING id`, name).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(ctx, `INSERT INTO alarm_limits (tag_id, hi, deadband) VALUES ($1, $2, 1)`, id, hi); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, q := range []string{"DELETE FROM measurements WHERE tag_id=$1", "DELETE FROM events WHERE tag_id=$1",
			"DELETE FROM alarm_limits WHERE tag_id=$1", "DELETE FROM tags WHERE id=$1"} {
			conn.Exec(ctx, q, id)
		}
	})
	return name
}

func findTag(t *testing.T, s *Store, name string) *Tag {
	tags, err := s.LoadTags(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, tg := range tags {
		if tg.Name == name {
			return tg
		}
	}
	t.Fatalf("LoadTags не вернул тег %s", name)
	return nil
}

func TestLoadTags(t *testing.T) {
	s, conn := testDB(t)
	name := testTag(t, conn, 42)
	tg := findTag(t, s, name)
	if tg.RegType != "holding" || tg.Address != 1 || tg.Scale != 0.5 || tg.Unit != "ед" {
		t.Errorf("поля тега: %+v", tg)
	}
	if tg.Limits == nil || tg.Limits.Hi == nil || *tg.Limits.Hi != 42 || tg.Limits.HiHi != nil {
		t.Errorf("уставки: %+v", tg.Limits)
	}
	for _, other := range []string{"reactor_temp", "gas_alarm"} {
		if findTag(t, s, other).Limits == nil {
			t.Errorf("у %s нет уставок", other)
		}
	}
}

func TestSaveSamples(t *testing.T) {
	s, conn := testDB(t)
	tg := findTag(t, s, testTag(t, conn, 42))
	ctx := context.Background()
	ts := time.Now().Truncate(time.Millisecond)
	if err := s.SaveSamples(ctx, ts, []Sample{{tg, 1.5}, {tg, 2.5}}); err != nil {
		t.Fatal(err)
	}
	var n int
	var sum float64
	var last time.Time
	err := conn.QueryRow(ctx, `SELECT count(*), sum(value), max(time) FROM measurements WHERE tag_id=$1`, tg.ID).Scan(&n, &sum, &last)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 || sum != 4 || !last.Equal(ts) {
		t.Errorf("в архиве: count=%d sum=%v time=%v", n, sum, last)
	}
	if err := s.SaveSamples(ctx, ts, nil); err != nil {
		t.Errorf("пустой пакет: %v", err)
	}
}

func TestSaveEvent(t *testing.T) {
	s, conn := testDB(t)
	tg := findTag(t, s, testTag(t, conn, 42))
	ctx := context.Background()
	if err := s.SaveEvent(ctx, time.Now(), tg, Hi, 43, "тест"); err != nil {
		t.Fatal(err)
	}
	var state, msg string
	var v float64
	err := conn.QueryRow(ctx, `SELECT state, value, message FROM events WHERE tag_id=$1`, tg.ID).Scan(&state, &v, &msg)
	if err != nil || state != "HI" || v != 43 || msg != "тест" {
		t.Errorf("событие: %s %v %q, %v", state, v, msg, err)
	}
}

func TestCurrentValues(t *testing.T) {
	s, conn := testDB(t)
	tg := findTag(t, s, testTag(t, conn, 42))
	ctx := context.Background()
	now := time.Now()
	s.SaveSamples(ctx, now.Add(-time.Second), []Sample{{tg, 1}})
	s.SaveSamples(ctx, now, []Sample{{tg, 2}})
	var v float64
	if err := conn.QueryRow(ctx, `SELECT value FROM current_values WHERE name=$1`, tg.Name).Scan(&v); err != nil || v != 2 {
		t.Errorf("current_values: %v, %v", v, err)
	}
}
