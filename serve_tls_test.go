package goss

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/goss-org/goss/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeSelfSignedCert writes a throwaway cert and key for 127.0.0.1 and
// returns their paths along with the parsed certificate.
func writeSelfSignedCert(t *testing.T) (string, string, *x509.Certificate) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "goss-test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	require.NoError(t, os.WriteFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600))
	require.NoError(t, os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), 0o600))
	return certPath, keyPath, cert
}

func TestServeTLSConfig(t *testing.T) {
	t.Parallel()
	certPath, keyPath, _ := writeSelfSignedCert(t)

	tests := map[string]struct {
		cert, key string
		wantTLS   bool
		wantErr   string
	}{
		"plain http":   {},
		"cert and key": {cert: certPath, key: keyPath, wantTLS: true},
		"cert only":    {cert: certPath, wantErr: "needs both"},
		"key only":     {key: keyPath, wantErr: "needs both"},
		"missing file": {cert: certPath, key: filepath.Join(t.TempDir(), "nope.pem"), wantErr: "loading tls key pair"},
		"swapped":      {cert: keyPath, key: certPath, wantErr: "loading tls key pair"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cfg, err := serveTLSConfig(&util.Config{ServeCertFile: tc.cert, ServeKeyFile: tc.key})
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
			if !tc.wantTLS {
				assert.Nil(t, cfg)
				return
			}
			require.NotNil(t, cfg)
			assert.Len(t, cfg.Certificates, 1)
			assert.Equal(t, uint16(tls.VersionTLS12), cfg.MinVersion)
		})
	}
}

func TestServeHealthOverTLS(t *testing.T) {
	t.Parallel()
	certPath, keyPath, cert := writeSelfSignedCert(t)

	config, err := util.NewConfig(
		util.WithSpecFile(filepath.Join("testdata", "passing.goss.yaml")),
		util.WithOutputFormat("json"),
	)
	require.NoError(t, err)
	config.ServeCertFile = certPath
	config.ServeKeyFile = keyPath

	tlsConfig, err := serveTLSConfig(config)
	require.NoError(t, err)
	handler, err := newHealthHandler(config)
	require.NoError(t, err)

	srv := httptest.NewUnstartedServer(handler)
	srv.TLS = tlsConfig
	srv.StartTLS()
	defer srv.Close()

	pool := x509.NewCertPool()
	pool.AddCert(cert)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}}}

	resp, err := client.Get(srv.URL + "/healthz")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, resp.TLS)
	assert.Equal(t, cert.Raw, resp.TLS.PeerCertificates[0].Raw)
}
