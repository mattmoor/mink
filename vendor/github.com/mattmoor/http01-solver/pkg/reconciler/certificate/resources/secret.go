/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License"); you
may not use this file except in compliance with the License.  You may
obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
implied.  See the License for the specific language governing
permissions and limitations under the License.
*/

package resources

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"knative.dev/pkg/kmeta"
	"knative.dev/serving/pkg/apis/networking/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

func IsValidCertificate(s *corev1.Secret, domains []string, minimumLifespan time.Duration) (bool, error) {
	if s.Data == nil {
		return false, nil
	}

	// TODO(mattmoor): Consider checking the private key as well, in case someone messed with it.

	// Crack open the certificate key.
	certPEM, ok := s.Data[corev1.TLSCertKey]
	if !ok {
		return false, nil
	}
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return false, fmt.Errorf("%q is not PEM encoded", corev1.TLSCertKey)
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return false, err
	}

	// Check whether all of the domains that we want covered are listed in the certificate.
	certDomains := sets.NewString(cert.DNSNames...)
	if !certDomains.HasAll(domains...) {
		return false, nil
	}

	// Compute the remaining useful lifespan of the certificate.
	lifespanLeft := cert.NotAfter.Sub(time.Now())

	// See if it is useful for at least our minimum.
	return lifespanLeft >= minimumLifespan, nil
}

// MakeSecret creates a TLS secret from the given certificate.
func MakeSecret(o *v1alpha1.Certificate, cert *tls.Certificate) (*corev1.Secret, error) {
	certPEM := &bytes.Buffer{}
	if err := pem.Encode(certPEM, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Leaf.Raw}); err != nil {
		return nil, err
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	if err != nil {
		return nil, err
	}
	privPEM := &bytes.Buffer{}
	if err := pem.Encode(privPEM, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return nil, err
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            o.Spec.SecretName,
			Namespace:       o.Namespace,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(o)},
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       certPEM.Bytes(),
			corev1.TLSPrivateKeyKey: privPEM.Bytes(),
		},
	}, nil
}
