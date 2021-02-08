package core

type Handler interface {
	Add(TCPConn)
	AddPacket(UDPPacket)
}
