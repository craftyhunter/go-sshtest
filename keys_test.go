package sshtest_test

import (
	"fmt"

	"github.com/craftyhunter/go-sshtest"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func ExampleNewSSHKeyPair() {
	// prepare keys
	privateKey, publicKey := sshtest.NewSSHKeyPair(2048)

	// auth in your client with agent
	myAgent := agent.NewKeyring()
	_ = myAgent.Add(agent.AddedKey{PrivateKey: privateKey})
	_ = &ssh.ClientConfig{
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(func() (signers []ssh.Signer, err error) {
				return myAgent.Signers()
			}),
		},
	}

	// auth in your client with signer from key
	signer, _ := ssh.NewSignerFromKey(privateKey)
	_ = &ssh.ClientConfig{
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(func() (signers []ssh.Signer, err error) {
				return []ssh.Signer{signer}, nil
			}),
		},
	}

	// publicKeyCallback in your server
	_ = &ssh.ServerConfig{
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			if string(publicKey.Marshal()) == string(pubKey.Marshal()) {
				return &ssh.Permissions{
					// Record the public key used for authentication.
					Extensions: map[string]string{
						"pubkey-fp": ssh.FingerprintSHA256(pubKey),
					},
				}, nil

			}
			return nil, fmt.Errorf("unknown public key for %q", c.User())
		},
	}
}
