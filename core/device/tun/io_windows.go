package tun

const (
	offset = 0

	defaultMTU = 0 /* auto */
)

func (t *TUN) Read(packet []byte) (int, error) {
	return t.nt.Read(packet, offset)
}

func (t *TUN) Write(packet []byte) (int, error) {
	return t.nt.Write(packet, offset)
}
