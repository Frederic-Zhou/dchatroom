package secret

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"io"

	"filippo.io/age"
)

var id *age.X25519Identity
var recipient *age.X25519Recipient
var pubkey ed25519.PublicKey
var prikey ed25519.PrivateKey
var recipients []string
var pubkeys []ed25519.PublicKey

func init() {
	id, recipient = genAge()
	prikey, pubkey = genEd25519()

	recipients = []string{
		recipient.String(),
	}
}

func genAge() (identity *age.X25519Identity, recipient *age.X25519Recipient) {
	id, _ := age.GenerateX25519Identity()
	return id, id.Recipient()
}

func genEd25519() (prikey ed25519.PrivateKey, pubkey ed25519.PublicKey) {
	pubkey, prikey, _ = ed25519.GenerateKey(rand.Reader)
	return
}

func GetLocalRecipient() string {
	return recipient.String()
}

func GetLocalPubKey() {

}

func AddRemoteRecipient(recipientStr string) []string {

	// recipient, err := age.ParseX25519Recipient(recipientStr)
	for _, v := range recipients {
		if v == recipientStr {
			return recipients
		}
	}
	recipients = append(recipients, recipientStr)
	return recipients

}

func AddRemotePubKeys(pubkey string) {

}

func Encrypt(planText string) (cryptoText string, err error) {

	recs := []age.Recipient{}

	for _, r := range recipients {
		rec, err := age.ParseX25519Recipient(r)
		if err == nil {
			recs = append(recs, rec)
		}
	}

	var dst bytes.Buffer

	w, err := age.Encrypt(&dst, recs...)
	if err != nil {
		return
	}
	if _, err = io.WriteString(w, planText); err != nil {
		return
	}
	err = w.Close()

	cryptoText = base64.RawStdEncoding.EncodeToString(dst.Bytes())

	return
}

func Decrypt(cryptoText string) (planText string, err error) {

	var dst bytes.Buffer

	cryptoBytes, err := base64.RawStdEncoding.DecodeString(cryptoText)
	if err != nil {
		return
	}

	_, err = dst.Write(cryptoBytes)
	if err != nil {
		return
	}

	r, err := age.Decrypt(&dst, id)
	if err != nil {
		return
	}

	out := &bytes.Buffer{}
	if _, err = io.Copy(out, r); err != nil {
		return
	}

	planText = out.String()

	return
}

func Sign() {

}

func Verify() {

}
