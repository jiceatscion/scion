// Copyright 2022 ETH Zurich
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

package trust

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/scionproto/scion/pkg/scrypto/cppki"
	"github.com/scionproto/scion/private/trust"
)

// TLSCertificateLoader is a wrapper for a SignerGen, converting the
// trust.Signer to an equivalent tls.Certificate.
type TLSCertificateLoader struct {
	SignerGen SignerGen
}

// GetCertificate returns the certificate representing the Signer generated by
// the SignerGen.
// This function can be bound to tls.Config.GetCertificate.
func (l TLSCertificateLoader) GetCertificate(
	hello *tls.ClientHelloInfo,
) (*tls.Certificate, error) {

	return l.Get(hello.Context())
}

// GetClientCertificate returns the certificate representing the Signer
// generated by the SignerGen.
// This function can be bound to tls.Config.GetClientCertificate.
func (l TLSCertificateLoader) GetClientCertificate(
	reqInfo *tls.CertificateRequestInfo,
) (*tls.Certificate, error) {

	return l.Get(reqInfo.Context())
}

// Get returns the certificate representing the Signer
// generated by the SignerGen.
func (l TLSCertificateLoader) Get(ctx context.Context) (*tls.Certificate, error) {
	signers, err := l.SignerGen.Generate(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	signer, err := trust.LastExpiring(signers, cppki.Validity{
		NotBefore: now,
		NotAfter:  now,
	})
	if err != nil {
		return nil, err
	}
	return toTLSCertificate(signer), nil
}

func toTLSCertificate(signer trust.Signer) *tls.Certificate {
	certificate := make([][]byte, len(signer.Chain))
	for i := range signer.Chain {
		certificate[i] = signer.Chain[i].Raw
	}
	return &tls.Certificate{
		Certificate: certificate,
		PrivateKey:  signer.PrivateKey,
		Leaf:        signer.Chain[0],
	}
}
