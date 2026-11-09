package main

type Tag struct {
	ID          int
	Name        string
	Description string
	Unit        string
	RegType     string
	Address     uint16
	Scale       float64
	Limits      *Limits
}

type Sample struct {
	Tag   *Tag
	Value float64
}

type State string

const (
	Normal State = "NORMAL"
	LoLo   State = "LO_LO"
	Lo     State = "LO"
	Hi     State = "HI"
	HiHi   State = "HI_HI"
	Alarm  State = "ALARM"
)

// nil - уставка не задана
type Limits struct {
	LoLo, Lo, Hi, HiHi *float64
	Deadband           float64
	AlarmValue         *float64 // дискретные теги
}
