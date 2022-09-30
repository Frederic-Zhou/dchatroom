package secret

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"sort"
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
	Accept   bool
	AKA      string
}

func init() {
	identity, recipient = genAge()
	prikey, pubkey = genEd25519()
	startRefreshRecipientsMap()

	recipientsMap.Store(recipient.String(), recipientItem{LastTime: time.Now(), Accept: true})

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

func StoreRemoteRecipient(recipientStr string, accept bool, aka string) {

	item, ok := recipientsMap.LoadOrStore(recipientStr, recipientItem{
		LastTime: time.Now(),
		Accept:   accept,
		AKA:      aka,
	})

	if ok { //存在，更新时间
		if i, o := item.(recipientItem); o {
			i.LastTime = time.Now()
			i.AKA = aka
			recipientsMap.Store(recipientStr, i)
		}
	}

}

func AcceptRemoteRecipient(recipientStr string, accept bool) {

	item, ok := recipientsMap.LoadOrStore(recipientStr, recipientItem{
		LastTime: time.Now(),
		Accept:   accept,
	})

	if ok { //存在，更新Accept
		if i, o := item.(recipientItem); o {
			i.Accept = accept
			recipientsMap.Store(recipientStr, i)
		}
	}
}

func DelRemoteRecipient(recipientStr string) {
	recipientsMap.Delete(recipientStr)
}

func GetRecipients() (recs []string) {

	recipientsMap.Range(func(key, value interface{}) bool {
		item, ok := value.(recipientItem)
		if ok {
			recs = append(recs, fmt.Sprintf("%s: %v,%s", key.(string), item.Accept, item.AKA))
		}
		return true
	})

	sort.Strings(recs)

	return
}

func AddRemotePubKeys(pubkey string) {

}

func Encrypt(planText string) (cryptoText string, err error) {

	recs := []age.Recipient{}

	recipientsMap.Range(func(key, value interface{}) bool {
		rec, err := age.ParseX25519Recipient(key.(string))
		if err == nil && value.(recipientItem).Accept {
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
