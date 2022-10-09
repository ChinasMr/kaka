package status

type Status uint32

const (
	// INIT The initial state, no valid SETUP has been received yet.
	INIT Status = iota

	// READY Last SETUP received was successful, reply sent or after
	// playing, last PAUSE received was successful, reply sent.
	READY

	// PLAYING Last PLAY received was successful, reply sent. Data is being sent.
	PLAYING

	// RECORDING The server is recording media data.
	RECORDING
)
