package status

type Status uint32

const (
	STARING Status = iota
	ANNOUNCED
	PRERECORD
	RECORD
	PLAY
)
