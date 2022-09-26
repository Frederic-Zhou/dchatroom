package test

import (
	"crypto/ed25519"
	"fmt"
	"testing"

	"filippo.io/age"
)

// This example demonstrates how to generate a
// collective signature involving two cosigners,
// and how to check the resulting collective signature.
func Example() {

	// Create keypairs for the two cosigners.
	pubKey1, priKey1, _ := ed25519.GenerateKey(nil)

	fmt.Println(len(pubKey1), pubKey1)
	fmt.Println(len(priKey1), priKey1)

	priKey2, _ := age.NewScryptIdentity("aaa")
	pubKey2, _ := age.NewScryptRecipient("aaa")

	fmt.Println(pubKey2)
	fmt.Println(priKey2)

}

func TestEd25518(t *testing.T) {
	Example()
}
