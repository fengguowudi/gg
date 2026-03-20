package trojan

import "testing"

func TestTrojanDialersUseSagerNet(t *testing.T) {
	t.Parallel()

	cases := []string{
		"trojan://pass@example.com:443?sni=real.example.com#tcp",
		"trojan-go://pass@example.com:443?type=ws&host=cdn.example.com&path=%2Fws&sni=real.example.com#ws",
		"trojan-go://pass@example.com:443?type=grpc&serviceName=svc&sni=real.example.com&encryption=ss%3Baes-128-gcm%3Bsecret#grpc-ss",
	}

	for _, link := range cases {
		link := link
		t.Run(link, func(t *testing.T) {
			t.Parallel()
			node, err := ParseTrojanURL(link)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := node.Dialer(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
