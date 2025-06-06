// Copyright 2021 Anapaya Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package trcs_test

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scionproto/scion/pkg/addr"
	"github.com/scionproto/scion/pkg/scrypto/cppki"
	"github.com/scionproto/scion/private/app/command"
	"github.com/scionproto/scion/scion-pki/certs"
	"github.com/scionproto/scion/scion-pki/key"
	"github.com/scionproto/scion/scion-pki/trcs"
)

func TestSign(t *testing.T) {
	outDir := t.TempDir()
	gen(t, outDir)

	testCases := map[string]struct {
		pld      string
		cert     string
		key      string
		signType string
	}{
		"sensitive - der|der": {
			pld:      filepath.Join(outDir, "ISD1-B1-S2.pld.der"),
			cert:     filepath.Join(outDir, "sensitive-voting.crt.der"),
			key:      filepath.Join(outDir, "sensitive-voting.key"),
			signType: "sensitive",
		},
		"sensitive - der|pem": {
			pld:      filepath.Join(outDir, "ISD1-B1-S2.pld.der"),
			cert:     filepath.Join(outDir, "sensitive-voting.crt.pem"),
			key:      filepath.Join(outDir, "sensitive-voting.key"),
			signType: "sensitive",
		},
		"sensitive - pem|der": {
			pld:      filepath.Join(outDir, "ISD1-B1-S2.pld.pem"),
			cert:     filepath.Join(outDir, "sensitive-voting.crt.der"),
			key:      filepath.Join(outDir, "sensitive-voting.key"),
			signType: "sensitive",
		},
		"sensitive - pem|pem": {
			pld:      filepath.Join(outDir, "ISD1-B1-S2.pld.pem"),
			cert:     filepath.Join(outDir, "sensitive-voting.crt.pem"),
			key:      filepath.Join(outDir, "sensitive-voting.key"),
			signType: "sensitive",
		},
		"regular": {
			pld:      filepath.Join(outDir, "ISD1-B1-S2.pld.pem"),
			cert:     filepath.Join(outDir, "regular-voting.crt.pem"),
			key:      filepath.Join(outDir, "regular-voting.key"),
			signType: "regular",
		},
		"root-ack": {
			pld:      filepath.Join(outDir, "ISD1-B1-S2.pld.pem"),
			cert:     filepath.Join(outDir, "cp-root.crt.pem"),
			key:      filepath.Join(outDir, "cp-root.key"),
			signType: "root-ack",
		},
		"sensitive-vote": {
			pld:      filepath.Join(outDir, "ISD1-B1-S2.pld.pem"),
			cert:     filepath.Join(outDir, "sensitive-voting.prev.crt.pem"),
			key:      filepath.Join(outDir, "sensitive-voting.prev.key"),
			signType: "sensitive-vote",
		},
		"regular-vote": {
			pld:      filepath.Join(outDir, "ISD1-B1-S2.pld.pem"),
			cert:     filepath.Join(outDir, "regular-voting.prev.crt.pem"),
			key:      filepath.Join(outDir, "regular-voting.prev.key"),
			signType: "regular-vote",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := trcs.RunSign(tc.pld, tc.cert, tc.key, "", "", outDir)
			assert.NoError(t, err)
			fn := filepath.Join(outDir, fmt.Sprintf("ISD1-B1-S2.1-1-%s.trc", tc.signType))

			_, err = trcs.DecodeFromFile(fn)
			assert.NoError(t, err)
		})
	}
}

func TestOpensslCompatible(t *testing.T) {
	if !strings.HasSuffix(os.Getenv("TEST_TARGET"), "go_integration_test") {
		t.Skip("This test only runs as integration test")
	}

	outDir := t.TempDir()
	gen(t, outDir)

	testCases := map[string]struct {
		pld      string
		cert     string
		key      string
		signType string
	}{
		"sensitive - der|pem": {
			pld:      filepath.Join(outDir, "ISD1-B1-S2.pld.der"),
			cert:     filepath.Join(outDir, "sensitive-voting.crt.pem"),
			key:      filepath.Join(outDir, "sensitive-voting.key"),
			signType: "sensitive",
		},
		"regular": {
			pld:      filepath.Join(outDir, "ISD1-B1-S2.pld.der"),
			cert:     filepath.Join(outDir, "regular-voting.crt.pem"),
			key:      filepath.Join(outDir, "regular-voting.key"),
			signType: "regular",
		},
		"root-ack": {
			pld:      filepath.Join(outDir, "ISD1-B1-S2.pld.der"),
			cert:     filepath.Join(outDir, "cp-root.crt.pem"),
			key:      filepath.Join(outDir, "cp-root.key"),
			signType: "root-ack",
		},
		"sensitive-vote": {
			pld:      filepath.Join(outDir, "ISD1-B1-S2.pld.der"),
			cert:     filepath.Join(outDir, "sensitive-voting.prev.crt.pem"),
			key:      filepath.Join(outDir, "sensitive-voting.prev.key"),
			signType: "sensitive-vote",
		},
		"regular-vote": {
			pld:      filepath.Join(outDir, "ISD1-B1-S2.pld.der"),
			cert:     filepath.Join(outDir, "regular-voting.prev.crt.pem"),
			key:      filepath.Join(outDir, "regular-voting.prev.key"),
			signType: "regular-vote",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := trcs.RunSign(tc.pld, tc.cert, tc.key, "", "", outDir)
			assert.NoError(t, err)
			fn := filepath.Join(outDir, fmt.Sprintf("ISD1-B1-S2.1-1-%s.trc", tc.signType))

			// Format TRC in DER for openssl.
			format := trcs.NewFormatCmd(command.StringPather("scion-pki"))
			format.SetArgs([]string{fn, "--format=der", "--out=" + fn + ".der"})
			err = format.Execute()
			require.NoError(t, err)

			cmd := exec.Command(
				"docker", "run", "--rm", "-v", outDir+":"+outDir, "emberstack/openssl",
				"openssl",
				"cms",
				"-verify",
				"-in="+fn+".der",
				"-inform=der",
				"-certfile="+tc.cert,
				"-CAfile="+tc.cert,
				"-purpose=any",
				"-no_check_time",
				">",
				"/dev/null",
			)
			out, err := cmd.CombinedOutput()
			require.NoError(t, err, string(out))
		})
	}
}

func gen(t *testing.T, outDir string) {
	// generate keys.
	sensitiveKey := genKey(t, filepath.Join(outDir, "sensitive-voting.key"))
	prevSensitiveKey := genKey(t, filepath.Join(outDir, "sensitive-voting.prev.key"))
	regularKey := genKey(t, filepath.Join(outDir, "regular-voting.key"))
	prevRegularKey := genKey(t, filepath.Join(outDir, "regular-voting.prev.key"))
	rootKey := genKey(t, filepath.Join(outDir, "cp-root.key"))

	// create self signed certificates.
	notBefore := time.Now().Add(-1 * time.Minute)
	notAfter := notBefore.Add(1 * time.Hour)
	sensitiveCert := genCert(t, cppki.Sensitive, sensitiveKey, notBefore, notAfter,
		filepath.Join(outDir, "sensitive-voting.crt"))
	regularCert := genCert(t, cppki.Regular, regularKey, notBefore, notAfter,
		filepath.Join(outDir, "regular-voting.crt"))
	rootCert := genCert(t, cppki.Root, rootKey, notBefore, notAfter,
		filepath.Join(outDir, "cp-root.crt"))
	genCert(t, cppki.Sensitive, prevSensitiveKey, notBefore, notAfter,
		filepath.Join(outDir, "sensitive-voting.prev.crt"))
	genCert(t, cppki.Regular, prevRegularKey, notBefore, notAfter,
		filepath.Join(outDir, "regular-voting.prev.crt"))

	trc := cppki.TRC{
		Version: 1,
		ID: cppki.TRCID{
			ISD:    addr.ISD(1),
			Serial: 2,
			Base:   1,
		},
		Validity: cppki.Validity{
			NotBefore: notBefore.Add(30 * time.Second),
			NotAfter:  notAfter.Add(-30 * time.Second),
		},
		CoreASes:          []addr.AS{1},
		AuthoritativeASes: []addr.AS{1},
		Quorum:            1,
		Description:       "This is a test TRC",
		Certificates:      []*x509.Certificate{sensitiveCert, regularCert, rootCert},
	}
	rawTRC, err := trc.Encode()
	require.NoError(t, err)
	encodedTRC := pem.EncodeToMemory(&pem.Block{
		Type:  "TRC PAYLOAD",
		Bytes: rawTRC,
	})
	require.NoError(t, os.WriteFile(filepath.Join(outDir, "ISD1-B1-S2.pld.der"),
		rawTRC, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(outDir, "ISD1-B1-S2.pld.pem"),
		encodedTRC, 0o644))
}

func genKey(t *testing.T, out string) key.PrivateKey {
	gen, err := key.GeneratePrivateKey("p256")
	require.NoError(t, err)
	encoded, err := key.EncodePEMPrivateKey(gen)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(out, encoded, 0o644))

	return gen
}

func genCert(
	t *testing.T,
	certType cppki.CertType,
	priv key.PrivateKey,
	notBefore, notAfter time.Time,
	out string,
) *x509.Certificate {
	certRaw, err := certs.CreateCertificate(certs.CertParams{
		Type: certType,
		Subject: pkix.Name{
			CommonName: "Anapaya Systems AG",
			ExtraNames: []pkix.AttributeTypeAndValue{
				{
					Type:  cppki.OIDNameIA,
					Value: "1-1",
				},
			},
		},
		PubKey:    priv.Public(),
		NotBefore: notBefore,
		NotAfter:  notAfter,
		CAKey:     priv,
	})
	require.NoError(t, err)
	encoded := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certRaw,
	})
	require.NotNil(t, encoded)
	cert, err := x509.ParseCertificate(certRaw)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(out+".der", encoded, 0o644))
	require.NoError(t, os.WriteFile(out+".pem", encoded, 0o644))

	return cert
}
