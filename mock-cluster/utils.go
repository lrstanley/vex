// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/rand/v2"
	"os"
	"strings"

	"github.com/go-faker/faker/v4"
)

func Must[T any](v T, err error) T {
	if err != nil {
		logger.Error("must", "error", err)
		os.Exit(1)
	}
	return v
}

func Must2[A, B any](a A, b B, err error) (A, B) { //nolint:gocritic
	if err != nil {
		logger.Error("must", "error", err)
		os.Exit(1)
	}
	return a, b
}

func Ptr[T any](v T) *T {
	return &v
}

func GetHostCertificates(addr string) []*x509.Certificate {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ //nolint:noctx
		InsecureSkipVerify: true, //nolint:gosec
	})
	if err != nil {
		logger.Error("failed to dial", "addr", addr, "error", err)
		os.Exit(1)
	}
	defer conn.Close()
	return conn.ConnectionState().PeerCertificates
}

func CertChainString(certs ...*x509.Certificate) string {
	chain := []string{}
	for _, cert := range certs {
		chain = append(chain, string(pem.EncodeToMemory(
			&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: cert.Raw,
			},
		)))
	}
	return strings.Join(chain, "\n")
}

func GenerateSlug(maxn int, sep string) (slug string) {
	for i := range rand.IntN(maxn) + 1 { //nolint:gosec
		if i > 0 {
			slug += sep
		}
		slug += faker.Word()
	}

	return slug
}
