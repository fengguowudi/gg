package hysteria2

import "testing"

func TestParseHysteria2URL(t *testing.T) {
	t.Parallel()

	node, err := ParseHysteria2URL("hysteria2://pass@example.com:123,5000-6000/?insecure=1&obfs=salamander&obfs-password=gawrgura&pinSHA256=deadbeef&sni=real.example.com#demo")
	if err != nil {
		t.Fatal(err)
	}
	if node.Server != "example.com" || node.Port != 123 {
		t.Fatalf("unexpected address: %s:%d", node.Server, node.Port)
	}
	if len(node.ServerPorts) != 1 || node.ServerPorts[0] != "5000:6000" {
		t.Fatalf("unexpected server ports: %#v", node.ServerPorts)
	}
	if node.Obfs != "salamander" || node.ObfsPassword != "gawrgura" {
		t.Fatalf("unexpected obfs config: %q %q", node.Obfs, node.ObfsPassword)
	}
	if !node.AllowInsecure {
		t.Fatal("expected insecure=true")
	}
	if _, err := node.Dialer(); err != nil {
		t.Fatal(err)
	}
}
