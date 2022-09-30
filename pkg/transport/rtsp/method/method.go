package method

type Method string

const (
	OPTIONS  Method = "OPTIONS"
	DESCRIBE Method = "DESCRIBE"
	ANNOUNCE Method = "ANNOUNCE"
	SETUP    Method = "SETUP"
	RECORD   Method = "RECORD"
)
