package sagernet

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	aTLS "github.com/sagernet/sing/common/tls"
)

type StdTLSConfig struct {
	config *tls.Config
}

func NewStdTLSConfig(serverName string, insecure bool, alpn []string, pinSHA256 string) (aTLS.Config, error) {
	cfg := &tls.Config{
		ServerName: serverName,
		NextProtos: append([]string(nil), alpn...),
	}
	if pinSHA256 == "" {
		cfg.InsecureSkipVerify = insecure
		return &StdTLSConfig{config: cfg}, nil
	}
	cfg.InsecureSkipVerify = true
	cfg.VerifyPeerCertificate = func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
		if len(rawCerts) == 0 {
			return fmt.Errorf("tls: no peer certificates")
		}
		certs := make([]*x509.Certificate, 0, len(rawCerts))
		for _, raw := range rawCerts {
			cert, err := x509.ParseCertificate(raw)
			if err != nil {
				return err
			}
			certs = append(certs, cert)
		}
		if !insecure {
			opts := x509.VerifyOptions{
				DNSName:       serverName,
				Intermediates: x509.NewCertPool(),
			}
			for _, cert := range certs[1:] {
				opts.Intermediates.AddCert(cert)
			}
			if _, err := certs[0].Verify(opts); err != nil {
				return err
			}
		}
		sum := sha256.Sum256(certs[0].Raw)
		if !matchesCertPin(sum[:], pinSHA256) {
			return fmt.Errorf("tls: certificate pin mismatch")
		}
		return nil
	}
	return &StdTLSConfig{config: cfg}, nil
}

func matchesCertPin(sum []byte, pin string) bool {
	pin = strings.TrimSpace(pin)
	if pin == "" {
		return true
	}
	hexPin := strings.ToLower(strings.ReplaceAll(pin, ":", ""))
	sumHex := hex.EncodeToString(sum)
	if hexPin == sumHex {
		return true
	}
	if pin == base64.StdEncoding.EncodeToString(sum) {
		return true
	}
	if pin == base64.RawStdEncoding.EncodeToString(sum) {
		return true
	}
	if pin == base64.URLEncoding.EncodeToString(sum) {
		return true
	}
	if pin == base64.RawURLEncoding.EncodeToString(sum) {
		return true
	}
	return false
}

func (c *StdTLSConfig) ServerName() string {
	return c.config.ServerName
}

func (c *StdTLSConfig) SetServerName(serverName string) {
	c.config.ServerName = serverName
}

func (c *StdTLSConfig) NextProtos() []string {
	return append([]string(nil), c.config.NextProtos...)
}

func (c *StdTLSConfig) SetNextProtos(nextProto []string) {
	c.config.NextProtos = append([]string(nil), nextProto...)
}

func (c *StdTLSConfig) STDConfig() (*aTLS.STDConfig, error) {
	return c.config.Clone(), nil
}

func (c *StdTLSConfig) Client(conn net.Conn) (aTLS.Conn, error) {
	return tls.Client(conn, c.config.Clone()), nil
}

func (c *StdTLSConfig) Clone() aTLS.Config {
	return &StdTLSConfig{config: c.config.Clone()}
}

func ClientHandshake(ctx context.Context, conn net.Conn, config aTLS.Config) (aTLS.Conn, error) {
	return aTLS.ClientHandshake(ctx, conn, config)
}
