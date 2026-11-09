package main

import (
	"os"
	"testing"
	"time"

	"github.com/simonvetter/modbus"
)

func plcURL(t *testing.T) string {
	u := os.Getenv("MODBUS_URL")
	if u == "" {
		t.Skip("MODBUS_URL не задан")
	}
	return u
}

func rawClient(t *testing.T, url string) *modbus.ModbusClient {
	c, err := modbus.NewClient(&modbus.ClientConfiguration{URL: url, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Open(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

func TestPLCReadTemperature(t *testing.T) {
	p := NewPLC(plcURL(t))
	v, err := p.Read(&Tag{RegType: "input", Address: 0, Scale: 0.1})
	if err != nil {
		t.Fatal(err)
	}
	if v < -50 || v > 200 {
		t.Errorf("температура вне диапазона: %.1f", v)
	}
}

func TestPLCHoldingRoundTrip(t *testing.T) {
	url := plcURL(t)
	c := rawClient(t, url)
	old, err := c.ReadRegister(0, modbus.HOLDING_REGISTER)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.WriteRegister(0, old) })

	if err := c.WriteRegister(0, 37); err != nil {
		t.Fatal(err)
	}
	p := NewPLC(url)
	sp, err := p.Read(&Tag{RegType: "holding", Address: 0, Scale: 1})
	if err != nil || sp != 37 {
		t.Fatalf("задание: %v, %v", sp, err)
	}
	time.Sleep(300 * time.Millisecond)
	act, err := p.Read(&Tag{RegType: "input", Address: 1, Scale: 1})
	if err != nil || act != 37 {
		t.Errorf("фактическая мощность: %v, %v", act, err)
	}
}

func TestPLCGasLeak(t *testing.T) {
	url := plcURL(t)
	c := rawClient(t, url)
	t.Cleanup(func() { c.WriteCoil(1, false) })
	p := NewPLC(url)
	gas := &Tag{RegType: "discrete", Address: 0, Scale: 1}
	for _, on := range []bool{true, false} {
		if err := c.WriteCoil(1, on); err != nil {
			t.Fatal(err)
		}
		time.Sleep(300 * time.Millisecond)
		v, err := p.Read(gas)
		want := 0.0
		if on {
			want = 1
		}
		if err != nil || v != want {
			t.Errorf("утечка=%v: датчик %v, %v", on, v, err)
		}
	}
}

func TestPLCUnknownType(t *testing.T) {
	p := NewPLC(plcURL(t))
	if _, err := p.Read(&Tag{RegType: "foo"}); err == nil {
		t.Error("ожидали ошибку для неизвестного типа регистра")
	}
}

func TestPLCNoServer(t *testing.T) {
	p := NewPLC("tcp://127.0.0.1:1")
	if _, err := p.Read(&Tag{RegType: "input"}); err == nil {
		t.Error("ожидали ошибку подключения")
	}
}
