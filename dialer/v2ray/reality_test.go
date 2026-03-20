package v2ray

import "testing"

func TestParseVlessRealityURL(t *testing.T) {
	t.Parallel()

	link := "vless://11111111-1111-1111-1111-111111111111@example.com:443?security=reality&type=tcp&sni=www.cloudflare.com&pbk=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&sid=6ba85179e30d4fc2&fp=chrome#demo"
	node, err := ParseVlessURL(link)
	if err != nil {
		t.Fatal(err)
	}
	if node.TLS != "reality" {
		t.Fatalf("unexpected security: %q", node.TLS)
	}
	if node.PublicKey == "" || node.ShortID == "" || node.Fingerprint != "chrome" {
		t.Fatalf("unexpected reality fields: %+v", node)
	}
	if node.Flow != "xtls-rprx-vision" {
		t.Fatalf("unexpected default flow: %q", node.Flow)
	}
	if _, err := node.Dialer(); err != nil {
		t.Fatal(err)
	}
}

func TestParseVlessRealityWSURL(t *testing.T) {
	t.Parallel()

	link := "vless://11111111-1111-1111-1111-111111111111@example.com:443?security=reality&type=ws&sni=www.cloudflare.com&host=cdn.example.com&path=%2Fws&pbk=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&sid=6ba85179e30d4fc2&fp=chrome#ws"
	node, err := ParseVlessURL(link)
	if err != nil {
		t.Fatal(err)
	}
	if node.Net != "ws" || node.Path != "/ws" || node.Host != "cdn.example.com" {
		t.Fatalf("unexpected ws reality fields: %+v", node)
	}
	if _, err := node.Dialer(); err != nil {
		t.Fatal(err)
	}
}

func TestParseVlessRealityGRPCURL(t *testing.T) {
	t.Parallel()

	link := "vless://11111111-1111-1111-1111-111111111111@example.com:443?security=reality&type=grpc&sni=www.cloudflare.com&serviceName=my-grpc&pbk=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&sid=6ba85179e30d4fc2&fp=chrome#grpc"
	node, err := ParseVlessURL(link)
	if err != nil {
		t.Fatal(err)
	}
	if node.Net != "grpc" || node.Path != "my-grpc" {
		t.Fatalf("unexpected grpc reality fields: %+v", node)
	}
	if _, err := node.Dialer(); err != nil {
		t.Fatal(err)
	}
}
