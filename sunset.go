package main

import (
	"crypto/x509"
	"time"
)

type sunsetSignatureAlgorithm struct {
	name string    // Human readable name of the signature algorithm.
	date time.Time // Date the signature algorithm will be sunset.
}

// sunsetSignatureAlgorithms is an algorithm to string mapping for certificate
// signature algorithms which have been or are being deprecated.  See the
// following links to learn more about SHA1's inclusion on this list.
// - https://technet.microsoft.com/en-us/library/security/2880823.aspx
// - http://googleonlinesecurity.blogspot.com/2014/09/gradually-sunsetting-sha-1.html
var sunsetSignatureAlgorithms = map[x509.SignatureAlgorithm]sunsetSignatureAlgorithm{
	x509.MD2WithRSA: {
		name: "MD2 with RSA",
		date: time.Now(),
	},
	x509.MD5WithRSA: {
		name: "MD5 with RSA",
		date: time.Now(),
	},
	x509.SHA1WithRSA: {
		name: "SHA1 with RSA",
		date: time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
	},
	x509.DSAWithSHA1: {
		name: "DSA with SHA1",
		date: time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
	},
	x509.ECDSAWithSHA1: {
		name: "ECDSA with SHA1",
		date: time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
	},
}
