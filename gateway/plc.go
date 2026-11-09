package main

// TODO 8: драйвер ПЛК на github.com/simonvetter/modbus.

type PLC struct {
	// TODO 8
}

func NewPLC(url string) *PLC {
	panic("TODO 8: NewPLC")
}

// Read читает тег и возвращает значение с учётом t.Scale.
func (p *PLC) Read(t *Tag) (float64, error) {
	panic("TODO 8: Read")
}
