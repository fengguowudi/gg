package trojan

import (
	"net/url"
	"testing"
)

func TestTrojanDialersUseSagerNet(t *testing.T) {
	t.Parallel()

	cases := []string{
		"trojan://pass@example.com:443?sni=real.example.com#tcp",
		"trojan-go://pass@example.com:443/ws?type=ws&host=cdn.example.com&sni=real.example.com#ws",
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

func TestParseTrojanGoWebSocketPathFromURLPath(t *testing.T) {
	t.Parallel()

	node, err := ParseTrojanURL("trojan-go://pass@example.com:443/ws?type=ws&host=cdn.example.com&sni=real.example.com#ws")
	if err != nil {
		t.Fatal(err)
	}
	if node.Protocol != "trojan-go" || node.Type != "ws" {
		t.Fatalf("unexpected trojan-go ws fields: %+v", node)
	}
	if node.Path != "/ws" {
		t.Fatalf("unexpected ws path: %q", node.Path)
	}
}

func TestParseTrojanGoWebSocketPathQueryFallback(t *testing.T) {
	t.Parallel()

	node, err := ParseTrojanURL("trojan-go://pass@example.com:443?type=ws&host=cdn.example.com&path=%2Flegacy&sni=real.example.com#ws")
	if err != nil {
		t.Fatal(err)
	}
	if node.Path != "/legacy" {
		t.Fatalf("unexpected fallback ws path: %q", node.Path)
	}
}

func TestExportTrojanGoWebSocketUsesURLPath(t *testing.T) {
	t.Parallel()

	link := (&Trojan{
		Name:     "ws",
		Server:   "example.com",
		Port:     443,
		Password: "pass",
		Sni:      "real.example.com",
		Type:     "ws",
		Host:     "cdn.example.com",
		Path:     "/ws",
		Protocol: "trojan-go",
	}).ExportToURL()

	u, err := url.Parse(link)
	if err != nil {
		t.Fatal(err)
	}
	if u.Path != "/ws" {
		t.Fatalf("unexpected exported path: %q", u.Path)
	}
	if got := u.Query().Get("path"); got != "" {
		t.Fatalf("unexpected redundant path query parameter: %q", got)
	}
}
