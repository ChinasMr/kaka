package status

type Status uint32

const (
	STARTING Status = iota
	ANNOUNCED
	PRERECORD
	RECORD
	PREPLAY
	PLAY
)
