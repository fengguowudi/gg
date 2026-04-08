package proxy

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNormalizeDNSServer(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "ipv4 no port", in: "1.1.1.1", want: "1.1.1.1:53"},
		{name: "hostname no port", in: "dns.google", want: "dns.google:53"},
		{name: "ipv4 with port", in: "1.1.1.1:5353", want: "1.1.1.1:5353"},
		{name: "ipv6 no port", in: "2606:4700:4700::1111", want: "[2606:4700:4700::1111]:53"},
		{name: "ipv6 with port", in: "[2606:4700:4700::1111]:5353", want: "[2606:4700:4700::1111]:5353"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := normalizeDNSServer(tc.in); got != tc.want {
				t.Fatalf("normalizeDNSServer(%q): got %q want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestProxyDNSTargetUsesConfiguredServer(t *testing.T) {
	t.Parallel()

	p := New(logrus.New(), nil, "8.8.8.8")
	if got := p.dnsTarget(DefaultDNSServer); got != "8.8.8.8:53" {
		t.Fatalf("unexpected dns target: got %q want %q", got, "8.8.8.8:53")
	}
}

func TestProxyDNSTargetFallsBackWhenUnset(t *testing.T) {
	t.Parallel()

	p := New(logrus.New(), nil, "")
	if got := p.dnsTarget(DefaultDNSServer); got != DefaultDNSServer {
		t.Fatalf("unexpected dns target: got %q want %q", got, DefaultDNSServer)
	}
}
