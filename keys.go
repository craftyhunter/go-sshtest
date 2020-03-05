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

func NewSSHKeyPair(bitSize int) (private ssh.Signer, public ssh.PublicKey) {
	key := NewRSAKey(bitSize)
	private, _ = ssh.NewSignerFromKey(key)
	public, _ = ssh.NewPublicKey(&key.PublicKey)
	return
}
