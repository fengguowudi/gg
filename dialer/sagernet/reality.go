package sagernet

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ed25519"
	stdtls "crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"reflect"
	"time"
	"unsafe"

	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	utls "github.com/metacubex/utls"
	aTLS "github.com/sagernet/sing/common/tls"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/net/http2"
)

type RealityTLSConfig struct {
	config    *utls.Config
	id        utls.ClientHelloID
	publicKey []byte
	shortID   [8]byte
}

func NewRealityTLSConfig(serverName, publicKey, shortID, fingerprint string, alpn []string) (aTLS.Config, error) {
	if serverName == "" {
		return nil, fmt.Errorf("missing server name")
	}
	id, err := uTLSClientHelloID(fingerprint)
	if err != nil {
		return nil, err
	}
	decodedKey, err := base64.RawURLEncoding.DecodeString(publicKey)
	if err != nil {
		return nil, fmt.Errorf("decode public key: %w", err)
	}
	if len(decodedKey) != 32 {
		return nil, fmt.Errorf("invalid public key")
	}
	var decodedShortID [8]byte
	if shortID != "" {
		n, err := hex.Decode(decodedShortID[:], []byte(shortID))
		if err != nil {
			return nil, fmt.Errorf("decode short id: %w", err)
		}
		if n > 8 {
			return nil, fmt.Errorf("invalid short id")
		}
	}
	cfg := &utls.Config{
		ServerName:             serverName,
		NextProtos:             append([]string(nil), alpn...),
		InsecureSkipVerify:     true,
		SessionTicketsDisabled: true,
	}
	return &RealityTLSConfig{
		config:    cfg,
		id:        id,
		publicKey: append([]byte(nil), decodedKey...),
		shortID:   decodedShortID,
	}, nil
}

func (c *RealityTLSConfig) ServerName() string {
	return c.config.ServerName
}

func (c *RealityTLSConfig) SetServerName(serverName string) {
	c.config.ServerName = serverName
}

func (c *RealityTLSConfig) NextProtos() []string {
	return append([]string(nil), c.config.NextProtos...)
}

func (c *RealityTLSConfig) SetNextProtos(nextProto []string) {
	if len(nextProto) == 1 && nextProto[0] == http2.NextProtoTLS {
		nextProto = append(nextProto, "http/1.1")
	}
	c.config.NextProtos = append([]string(nil), nextProto...)
}

func (c *RealityTLSConfig) STDConfig() (*aTLS.STDConfig, error) {
	return nil, fmt.Errorf("unsupported usage for reality")
}

func (c *RealityTLSConfig) Client(conn net.Conn) (aTLS.Conn, error) {
	return c.ClientHandshake(context.Background(), conn)
}

func (c *RealityTLSConfig) ClientHandshake(ctx context.Context, conn net.Conn) (aTLS.Conn, error) {
	verifier := &realityVerifier{
		serverName: c.config.ServerName,
	}
	uConfig := c.config.Clone()
	uConfig.InsecureSkipVerify = true
	uConfig.SessionTicketsDisabled = true
	uConfig.VerifyPeerCertificate = verifier.VerifyPeerCertificate

	uConn := utls.UClient(conn, uConfig, c.id)
	verifier.UConn = uConn
	if err := uConn.BuildHandshakeState(); err != nil {
		return nil, err
	}
	for _, extension := range uConn.Extensions {
		if ce, ok := extension.(*utls.SupportedCurvesExtension); ok {
			filtered := ce.Curves[:0]
			for _, curveID := range ce.Curves {
				if curveID != utls.X25519MLKEM768 {
					filtered = append(filtered, curveID)
				}
			}
			ce.Curves = filtered
		}
		if ks, ok := extension.(*utls.KeyShareExtension); ok {
			filtered := ks.KeyShares[:0]
			for _, share := range ks.KeyShares {
				if share.Group != utls.X25519MLKEM768 {
					filtered = append(filtered, share)
				}
			}
			ks.KeyShares = filtered
		}
		if alpnExtension, ok := extension.(*utls.ALPNExtension); ok && len(uConfig.NextProtos) > 0 {
			alpnExtension.AlpnProtocols = uConfig.NextProtos
		}
	}
	if err := uConn.BuildHandshakeState(); err != nil {
		return nil, err
	}

	hello := uConn.HandshakeState.Hello
	hello.SessionId = make([]byte, 32)
	copy(hello.Raw[39:], hello.SessionId)
	hello.SessionId[0] = 1
	hello.SessionId[1] = 8
	hello.SessionId[2] = 1
	binary.BigEndian.PutUint32(hello.SessionId[4:], uint32(time.Now().Unix()))
	copy(hello.SessionId[8:], c.shortID[:])

	publicKey, err := ecdh.X25519().NewPublicKey(c.publicKey)
	if err != nil {
		return nil, err
	}
	keyShareKeys := uConn.HandshakeState.State13.KeyShareKeys
	if keyShareKeys == nil || keyShareKeys.Ecdhe == nil {
		return nil, fmt.Errorf("nil ecdhe key")
	}
	authKey, err := keyShareKeys.Ecdhe.ECDH(publicKey)
	if err != nil {
		return nil, err
	}
	verifier.authKey = authKey
	if _, err = hkdf.New(sha256.New, authKey, hello.Random[:20], []byte("REALITY")).Read(authKey); err != nil {
		return nil, err
	}
	aesBlock, _ := aes.NewCipher(authKey)
	aesGCM, _ := cipher.NewGCM(aesBlock)
	aesGCM.Seal(hello.SessionId[:0], hello.Random[20:], hello.SessionId[:16], hello.Raw)
	copy(hello.Raw[39:], hello.SessionId)

	if err = uConn.HandshakeContext(ctx); err != nil {
		return nil, err
	}
	if !verifier.verified {
		return nil, fmt.Errorf("reality verification failed")
	}
	return &RealityClientConn{UConn: uConn}, nil
}

func (c *RealityTLSConfig) Clone() aTLS.Config {
	clonedKey := append([]byte(nil), c.publicKey...)
	return &RealityTLSConfig{
		config:    c.config.Clone(),
		id:        c.id,
		publicKey: clonedKey,
		shortID:   c.shortID,
	}
}

type realityVerifier struct {
	*utls.UConn
	serverName string
	authKey    []byte
	verified   bool
}

func (c *realityVerifier) VerifyPeerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	field, _ := reflect.TypeOf(c.Conn).Elem().FieldByName("peerCertificates")
	certs := *(*([]*x509.Certificate))(unsafe.Pointer(uintptr(unsafe.Pointer(c.Conn)) + field.Offset))
	if len(certs) == 0 {
		return fmt.Errorf("missing peer certificates")
	}
	if pub, ok := certs[0].PublicKey.(ed25519.PublicKey); ok {
		h := hmac.New(sha512.New, c.authKey)
		_, _ = h.Write(pub)
		if bytes.Equal(h.Sum(nil), certs[0].Signature) {
			c.verified = true
			return nil
		}
	}
	opts := x509.VerifyOptions{
		DNSName:       c.serverName,
		Intermediates: x509.NewCertPool(),
	}
	for _, cert := range certs[1:] {
		opts.Intermediates.AddCert(cert)
	}
	if _, err := certs[0].Verify(opts); err != nil {
		return err
	}
	return nil
}

type RealityClientConn struct {
	*utls.UConn
}

func (c *RealityClientConn) ConnectionState() stdtls.ConnectionState {
	state := c.Conn.ConnectionState()
	return stdtls.ConnectionState{
		Version:                     state.Version,
		HandshakeComplete:           state.HandshakeComplete,
		DidResume:                   state.DidResume,
		CipherSuite:                 state.CipherSuite,
		NegotiatedProtocol:          state.NegotiatedProtocol,
		NegotiatedProtocolIsMutual:  state.NegotiatedProtocolIsMutual,
		ServerName:                  state.ServerName,
		PeerCertificates:            state.PeerCertificates,
		VerifiedChains:              state.VerifiedChains,
		SignedCertificateTimestamps: state.SignedCertificateTimestamps,
		OCSPResponse:                state.OCSPResponse,
		TLSUnique:                   state.TLSUnique,
	}
}

func (c *RealityClientConn) Upstream() any {
	return c.UConn
}

func (c *RealityClientConn) CloseWrite() error {
	return c.Close()
}

func (c *RealityClientConn) ReaderReplaceable() bool {
	return true
}

func (c *RealityClientConn) WriterReplaceable() bool {
	return true
}

var (
	randomFingerprint     utls.ClientHelloID
	randomizedFingerprint utls.ClientHelloID
)

func init() {
	modernFingerprints := []utls.ClientHelloID{
		utls.HelloChrome_Auto,
		utls.HelloFirefox_Auto,
		utls.HelloEdge_Auto,
		utls.HelloSafari_Auto,
		utls.HelloIOS_Auto,
	}
	randomFingerprint = modernFingerprints[rand.Intn(len(modernFingerprints))]

	weights := utls.DefaultWeights
	weights.TLSVersMax_Set_VersionTLS13 = 1
	weights.FirstKeyShare_Set_CurveP256 = 0
	randomizedFingerprint = utls.HelloRandomized
	randomizedFingerprint.Seed, _ = utls.NewPRNGSeed()
	randomizedFingerprint.Weights = &weights
}

func uTLSClientHelloID(name string) (utls.ClientHelloID, error) {
	switch name {
	case "chrome_psk", "chrome_psk_shuffle", "chrome_padding_psk_shuffle", "chrome_pq", "chrome_pq_psk":
		fallthrough
	case "chrome", "":
		return utls.HelloChrome_Auto, nil
	case "firefox":
		return utls.HelloFirefox_Auto, nil
	case "edge":
		return utls.HelloEdge_Auto, nil
	case "safari":
		return utls.HelloSafari_Auto, nil
	case "360":
		return utls.Hello360_Auto, nil
	case "qq":
		return utls.HelloQQ_Auto, nil
	case "ios":
		return utls.HelloIOS_Auto, nil
	case "android":
		return utls.HelloAndroid_11_OkHttp, nil
	case "random":
		return randomFingerprint, nil
	case "randomized":
		return randomizedFingerprint, nil
	default:
		return utls.ClientHelloID{}, fmt.Errorf("unknown uTLS fingerprint: %s", name)
	}
}
