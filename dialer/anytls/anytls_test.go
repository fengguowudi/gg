package anytls

import "testing"

func TestParseAnyTLSURL(t *testing.T) {
	t.Parallel()

	node, err := ParseAnyTLSURL("anytls://letmein@example.com/?sni=real.example.com#demo")
	if err != nil {
		t.Fatal(err)
	}
	if node.Server != "example.com" || node.Port != 443 {
		t.Fatalf("unexpected address: %s:%d", node.Server, node.Port)
	}
	if node.Password != "letmein" {
		t.Fatalf("unexpected password: %q", node.Password)
	}
	if node.SNI != "real.example.com" {
		t.Fatalf("unexpected sni: %q", node.SNI)
	}
	if _, err := node.Dialer(); err != nil {
		t.Fatal(err)
	}
}
