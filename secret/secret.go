package secret

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"io"
	"sync"
	"time"

	"filippo.io/age"
)

var identity *age.X25519Identity
var recipient *age.X25519Recipient
var pubkey ed25519.PublicKey
var prikey ed25519.PrivateKey
var recipientsMap sync.Map
var pubkeys []ed25519.PublicKey

type recipientItem struct {
	LastTime time.Time
}

func init() {
	identity, recipient = genAge()
	prikey, pubkey = genEd25519()
	startRefreshRecipientsMap()

	recipientsMap.Store(recipient.String(), recipientItem{LastTime: time.Now()})

}

func startRefreshRecipientsMap() {

	ctx := context.Background()
	go func(ctx context.Context) {

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				recipientsMap.Range(func(k interface{}, v interface{}) bool {
					item, ok := v.(recipientItem)
					if ok && item.LastTime.Add(15*time.Second).Before(time.Now()) {
						recipientsMap.Delete(k)
					}
					return true
				})
			}
		}

	}(ctx)

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

func GetLocalPubKey() string {
	return base64.RawStdEncoding.EncodeToString(pubkey)
}

func StoreRemoteRecipient(recipientStr string) {
	recipientsMap.Store(recipientStr, recipientItem{
		LastTime: time.Now(),
	})
}

func DelRemoteRecipient(recipientStr string) {
	recipientsMap.Delete(recipientStr)
}

func RecipientsCount() int {
	count := 0
	recipientsMap.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

func AddRemotePubKeys(pubkey string) {

}

func Encrypt(planText string) (cryptoText string, err error) {

	recs := []age.Recipient{}

	recipientsMap.Range(func(key, value interface{}) bool {
		rec, err := age.ParseX25519Recipient(key.(string))
		if err == nil {
			recs = append(recs, rec)
		}
		return true
	})

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

	r, err := age.Decrypt(&dst, identity)
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
