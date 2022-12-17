package netx

import "net"

// InternalIp 获取内网IP。
func InternalIp() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, v := range interfaces {
		if isEthDown(v.Flags) || isLoopback(v.Flags) {
			continue
		}

		addrs, err := v.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					return ipNet.IP.String()
				}
			}
		}
	}

	return ""
}

func isEthDown(f net.Flags) bool {
	return f&net.FlagUp != net.FlagUp
}

func isLoopback(f net.Flags) bool {
	return f&net.FlagLoopback == net.FlagLoopback
}
