package listener

import "net"

type rtspListener struct {
	netListener *net.TCPListener
}
