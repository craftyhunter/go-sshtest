package sshtest

import (
	"crypto/rand"
	"crypto/rsa"
)

func generateRSAKey(bitSize int) (key *rsa.PrivateKey, err error) {
	// generate key
	return rsa.GenerateKey(rand.Reader, bitSize)
}
