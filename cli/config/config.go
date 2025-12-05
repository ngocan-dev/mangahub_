type Config struct {
	GRPCServer string `yaml:"grpc_server"`
	HTTPServer string `yaml:"http_server"`
	TCPServer  string `yaml:"tcp_server"`
	UDPServer  string `yaml:"udp_server"`
	WSServer   string `yaml:"ws_server"`
	Token      string `yaml:"token"`
}