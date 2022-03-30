package engine

type Key struct {
	MTU        int    `yaml:"mtu"`
	Mark       int    `yaml:"fwmark"`
	UDPTimeout int    `yaml:"udp-timeout"`
	Proxy      string `yaml:"proxy"`
	RestAPI    string `yaml:"restapi"`
	Device     string `yaml:"device"`
	LogLevel   string `yaml:"loglevel"`
	Interface  string `yaml:"interface"`
}
