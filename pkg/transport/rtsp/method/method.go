package method

type Method string

func (m Method) String() string {
	return string(m)
}

const (
	OPTIONS  Method = "OPTIONS"
	DESCRIBE Method = "DESCRIBE"
	ANNOUNCE Method = "ANNOUNCE"
	SETUP    Method = "SETUP"
	RECORD   Method = "RECORD"
	TEARDOWN Method = "TEARDOWN"
	PLAY     Method = "PLAY"
	PAUSE    Method = "PAUSE"
)
