package main

import (
	"strings"
	"testing"
)

func ptr(v float64) *float64 { return &v }

func TestEvaluate(t *testing.T) {
	analog := Limits{LoLo: ptr(5), Lo: ptr(10), Hi: ptr(80), HiHi: ptr(90), Deadband: 2}
	cases := []struct {
		name string
		prev State
		v    float64
		want State
	}{
		{"норма", Normal, 50, Normal},
		{"ровно H", Normal, 80, Hi},
		{"чуть ниже H без аварии", Normal, 79.9, Normal},
		{"H держится в зоне возврата", Hi, 78.5, Hi},
		{"H снимается за зоной", Hi, 77.9, Normal},
		{"ровно H-deadband уже норма", Hi, 78, Normal},
		{"сразу HH", Normal, 95, HiHi},
		{"HH держится в зоне", HiHi, 88.5, HiHi},
		{"из HH в H", HiHi, 87.5, Hi},
		{"из HH сразу в норму", HiHi, 70, Normal},
		{"ровно L", Normal, 10, Lo},
		{"L держится в зоне", Lo, 11.5, Lo},
		{"L снимается", Lo, 12.1, Normal},
		{"LL", Normal, 4, LoLo},
		{"из LL в L", LoLo, 7.5, Lo},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Evaluate(c.prev, c.v, analog); got != c.want {
				t.Errorf("Evaluate(%s, %.1f) = %s, ожидали %s", c.prev, c.v, got, c.want)
			}
		})
	}
}

func TestEvaluateOnlyHigh(t *testing.T) {
	l := Limits{Hi: ptr(80), Deadband: 1}
	if got := Evaluate(Normal, -100, l); got != Normal {
		t.Errorf("без нижних уставок получили %s", got)
	}
}

func TestEvaluateDiscrete(t *testing.T) {
	l := Limits{AlarmValue: ptr(1)}
	for _, c := range []struct {
		prev State
		v    float64
		want State
	}{{Normal, 1, Alarm}, {Alarm, 1, Alarm}, {Alarm, 0, Normal}, {Normal, 0, Normal}} {
		if got := Evaluate(c.prev, c.v, l); got != c.want {
			t.Errorf("Evaluate(%s, %v) = %s, ожидали %s", c.prev, c.v, got, c.want)
		}
	}
}

func TestMessage(t *testing.T) {
	l := Limits{Hi: ptr(80), HiHi: ptr(90)}
	msg := Message("Температура", "°C", HiHi, 91.3, l)
	for _, want := range []string{"Температура", "90.0", "91.3 °C"} {
		if !strings.Contains(msg, want) {
			t.Errorf("в сообщении %q нет %q", msg, want)
		}
	}
	if msg := Message("Температура", "°C", Normal, 70, l); !strings.Contains(msg, "норм") {
		t.Errorf("сообщение о возврате в норму: %q", msg)
	}
	if msg := Message("Датчик газа", "", Normal, 0, Limits{AlarmValue: ptr(1)}); strings.Contains(msg, " )") {
		t.Errorf("лишний пробел при пустой единице: %q", msg)
	}
}
