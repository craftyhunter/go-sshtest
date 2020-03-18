package sshtest

import (
	"crypto/rand"
	"crypto/rsa"

	"golang.org/x/crypto/ssh"
)

func NewRSAKey(bitSize int) *rsa.PrivateKey {
	key, _ := rsa.GenerateKey(rand.Reader, bitSize)
	return key
}

// NewSSHKeyPair generate new key pair
func NewSSHKeyPair(bitSize int) (private *rsa.PrivateKey, public ssh.PublicKey) {
	private = NewRSAKey(bitSize)
	public, _ = ssh.NewPublicKey(&private.PublicKey)
	return
}
