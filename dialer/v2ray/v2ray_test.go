package v2ray

import "testing"

func TestNonRealityDialersUseSagerNet(t *testing.T) {
	t.Parallel()

	cases := []*V2Ray{
		{
			Ps:       "vmess-ws",
			Add:      "example.com",
			Port:     "443",
			ID:       "11111111-1111-1111-1111-111111111111",
			Aid:      "0",
			Net:      "ws",
			Host:     "cdn.example.com",
			Path:     "/ws",
			TLS:      "tls",
			Protocol: "vmess",
		},
		{
			Ps:       "vless-grpc",
			Add:      "example.com",
			Port:     "443",
			ID:       "11111111-1111-1111-1111-111111111111",
			Net:      "grpc",
			Path:     "grpc-service",
			SNI:      "real.example.com",
			Flow:     "xtls-rprx-direct",
			Protocol: "vless",
		},
		{
			Ps:       "vmess-h2",
			Add:      "example.com",
			Port:     "443",
			ID:       "11111111-1111-1111-1111-111111111111",
			Aid:      "0",
			Net:      "h2",
			Host:     "h2.example.com",
			Path:     "/ray",
			TLS:      "tls",
			Protocol: "vmess",
		},
	}

	for _, node := range cases {
		node := node
		t.Run(node.Ps, func(t *testing.T) {
			t.Parallel()
			if _, err := node.Dialer(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
