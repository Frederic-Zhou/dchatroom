package test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"testing"

	"filippo.io/age"
)

var out bytes.Buffer

func TestEncrypto(t *testing.T) {
	// publicKey := "age106wz8r7gm9glr2lza8tf82zzk5cdnxg37zjty9hnp5dlh5ghpp7saz7kxu"
	// recipient, err := age.ParseX25519Recipient(publicKey)
	// if err != nil {
	// 	log.Fatalf("Failed to parse public key %q: %v", publicKey, err)
	// }

	recipient, err := age.NewScryptRecipient("123")
	fmt.Println("err:", err)

	w, err := age.Encrypt(&out, recipient)
	if err != nil {
		log.Fatalf("Failed to create encrypted file: %v", err)
	}
	if _, err := io.WriteString(w, "Black lives matter."); err != nil {
		log.Fatalf("Failed to write to encrypted file: %v", err)
	}
	if err := w.Close(); err != nil {
		log.Fatalf("Failed to close encrypted file: %v", err)
	}

	fmt.Printf("Encrypted file size: %d\n %v\n", out.Len(), out.String())
}

func TestDecrypto(t *testing.T) {

	identity, err := age.NewScryptIdentity("123")
	// privateKey := "AGE-SECRET-KEY-1R0G0RWWJ690HX299J9YW2ADQK08FYSK08QW2KG6CHV7HD3SMQS8QXJ3NV7"
	// identity, err := age.ParseX25519Identity(privateKey)

	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	r, err := age.Decrypt(&out, identity)
	if err != nil {
		log.Fatalf("Failed to open encrypted file: %v", err)
	}
	out := &bytes.Buffer{}
	if _, err := io.Copy(out, r); err != nil {
		log.Fatalf("Failed to read encrypted file: %v", err)
	}

	fmt.Printf("File contents: %q\n", out.Bytes())
}

func TestCreate(t *testing.T) {
	id, _ := age.GenerateX25519Identity()
	fmt.Println(id.String())
	fmt.Println(id.Recipient().String())
}
