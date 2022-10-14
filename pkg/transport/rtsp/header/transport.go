package header

import (
	"fmt"
	"strconv"
	"strings"
)

const defaultTrans = "RTP/AVP"
const LowerTransTCP = "RTP/AVP/TCP"
const LoweTransUDP = "RTP/AVP/UDP"
const ParamMulticast = "multicast"
const ParamUnicast = "unicast"
const paramInterleaved = "interleaved"
const modeRecord = "mode=record"
const clientPort = "client_port"
const serverPort = "server_port"

type TransportHeader map[string]struct{}

func NewTransportHeader(lt string, param ...string) string {
	ks := make([]string, 0)
	ks = append(ks, lt)
	ks = append(ks, param...)
	return strings.Join(ks, ";")
}

func NewInterleavedParam(ch1 int64, ch2 int64) string {
	return fmt.Sprintf("%s=%d-%d", paramInterleaved, ch1, ch2)
}

func NewClientPort(p1 int64, p2 int64) string {
	return fmt.Sprintf("%s=%d-%d", clientPort, p1, p2)
}

func NewServerPort(p1 int64, p2 int64) string {
	return fmt.Sprintf("%s=%d-%d", serverPort, p1, p2)
}

func (t TransportHeader) Has(key string) bool {
	_, ok := t[key]
	return ok
}

func (t TransportHeader) Value(k string) string {
	prefix := fmt.Sprintf("%s=", k)
	for key := range t {
		if strings.HasPrefix(key, prefix) {
			return key[len(prefix):]
		}
	}
	return ""
}

func (t TransportHeader) LowerTransportTCP() bool {
	if t.Has(LowerTransTCP) {
		return true
	}
	return false
}

func (t TransportHeader) Multicast() bool {
	if t.Has(ParamMulticast) {
		return true
	}
	return false
}

func (t TransportHeader) Interleaved() (int64, int64, bool) {
	v := t.Value(paramInterleaved)
	if v == "" {
		return 0, 0, false
	}
	// channel = 1*3(DIGIT)
	ps := strings.Split(v, "-")
	if len(ps) != 2 {
		return 0, 0, false
	}
	c1, err := strconv.ParseInt(ps[0], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	c2, err := strconv.ParseInt(ps[1], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	return c1, c2, true
}

func (t TransportHeader) ClientPort() (int64, int64, bool) {
	v := t.Value(clientPort)
	if v == "" {
		return 0, 0, false
	}
	ps := strings.Split(v, "-")
	if len(ps) != 2 {
		return 0, 0, false
	}
	c1, err := strconv.ParseInt(ps[0], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	c2, err := strconv.ParseInt(ps[1], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	return c1, c2, true
}

func (t TransportHeader) Record() bool {
	return t.Has(modeRecord)
}

func (t TransportHeader) Validate() bool {
	if !t.Has(defaultTrans) && !t.Has(LoweTransUDP) && !t.Has(LowerTransTCP) {
		return false
	}
	if !t.Has(ParamUnicast) && !t.Has(ParamMulticast) {
		return false
	}
	return true
}
