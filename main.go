package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"myd/ipfsapi"
	"myd/secret"
	"myd/view"
	"os"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
)

var currentTopic string = ""
var currentSubChan chan []byte
var cancelSub context.CancelFunc
var currentAKA, _ = os.Hostname()
var autoAccept *bool

func main() {

	// ipfsapi.Version()

	// ipfsapi.SubLs()
	// ipfsapi.SubPeers("hellodawngrp")
	// ipfsapi.Pub("hellodawngrp", "hello")
	// c, _ := ipfsapi.Sub("hellodawngrp")
	// for line := range c {
	// 	fmt.Print(string(line))
	// }

	autoAccept = flag.Bool("a", true, "auto accept, Default: true")
	flag.Parse()

	_, err := ipfsapi.Version()
	if err != nil {
		fmt.Println(`
		=============================
		||run 'ipfs daemon' first!!||
		=============================
		`)
		return
	}
	view.Run(commandHandler)
}

func commandHandler(text string) {

	// view.AddMessage([]byte(currentTopic + "/" + text))
	switch {
	case strings.HasPrefix(text, "/sub "):
		//取得订阅的Topic，去掉前后的空白
		topic := strings.TrimSpace(strings.TrimPrefix(text, "/sub "))
		if topic == "" {
			return
		}

		//将之前的ctx cancel
		if cancelSub != nil {
			cancelSub()
		}

		//创建一个新的可取消的ctx 和 cancelSub函数
		var ctx context.Context
		ctx, cancelSub = context.WithCancel(context.Background())

		//取得订阅的chan
		var err error
		currentSubChan, err = ipfsapi.Sub(ctx, topic)
		if err != nil {
			view.SetInfoView("[red]sub error:[white]" + err.Error())
			return
		}

		//设置当前Topic为当前sub命令的Topic
		currentTopic = topic

		view.SetInfoView(fmt.Sprintf("[black]Topic:%s, AKA:%s, Peers:%s", currentTopic, currentAKA, "0"))
		//定时刷新topic信息
		go infoAndHeartBitHandler(ctx, topic)
		//开goroutin不断获取ipfsapi sub的通道数据
		go messageHandler(ctx, currentSubChan)
		//订阅发个heartbit
		pubHandler("/heartbit", false)
		pubHandler(text, false)

	case strings.HasPrefix(text, "/aka "):
		aka := strings.TrimSpace(strings.TrimPrefix(text, "/aka "))
		if aka == "" {
			return
		}
		currentAKA = aka
		pubHandler(text, false)
	case strings.HasPrefix(text, "/accept "):
		accept := strings.TrimSpace(strings.TrimPrefix(text, "/accept "))
		if accept == "" {
			return
		}
		secret.AcceptRemoteRecipient(accept, true)
	case strings.HasPrefix(text, "/reject "):
		reject := strings.TrimSpace(strings.TrimPrefix(text, "/reject "))
		if reject == "" {
			return
		}
		secret.AcceptRemoteRecipient(reject, false)
	default:
		pubHandler(text, true)
	}

}

func pubHandler(text string, toEncrypt bool) {
	// view.AddMessage([]byte("[red]pub plantext:[white]" + text))
	var err error
	if currentTopic == "" {
		view.SetInfoView("[red]sub a topic first[white]")
		return
	}

	if text == "" {
		return
	}

	defer func() {
		if err != nil {
			view.AddMessage([]byte("[red]pub error:[white]" + err.Error()))
		}
	}()

	if toEncrypt {
		text, err = secret.Encrypt(text)
		if err != nil {
			view.AddMessage([]byte("Encrypt Error:" + err.Error()))
			return
		}
	}

	data, _ := json.Marshal(Data{
		Text:      text,
		AKA:       currentAKA,
		PubKey:    secret.GetLocalPubKey(),
		Recipient: secret.GetLocalRecipient(),
		Encrypted: toEncrypt,
	})

	_, err = ipfsapi.Pub(currentTopic, string(data))
}

func messageHandler(ctx context.Context, c chan []byte) {

	for {
		select {
		case <-ctx.Done():
			return
		case line := <-c:

			// view.AddMessage(line)
			message := Message{}
			if err := json.Unmarshal(line, &message); err != nil {
				view.AddMessage([]byte("json1:" + err.Error()))
				return
			}

			topicIDs := []string{}
			for _, t := range message.TopicIDs {
				topicIDs = append(topicIDs, string(ipfsapi.Base64urlDecode(t)))
			}
			seqnoBytes := ipfsapi.Base64urlDecode(message.Seqno)
			seqno := binary.BigEndian.Uint64(seqnoBytes)

			data := ipfsapi.Base64urlDecode(message.Data)
			dataObj := Data{}

			if err := json.Unmarshal(data, &dataObj); err != nil {
				view.AddMessage([]byte("json2:" + err.Error() + "\n" + string(data)))
				return
			}

			// recs := secret.GetRecipients()
			//todo: check signtrue
			if dataObj.Encrypted {
				var err error
				if dataObj.Text, err = secret.Decrypt(dataObj.Text); err != nil {
					dataObj.Text = fmt.Sprintf("Decrypt Error:%s\n%s", err.Error(), dataObj.Text)
				}
			}

			switch {
			case strings.HasPrefix(dataObj.Text, "/heartbit"): //收到心跳
				secret.StoreRemoteRecipient(dataObj.Recipient, *autoAccept, dataObj.AKA)
				//do nothing
			case strings.HasPrefix(dataObj.Text, "/sub"):
				pubHandler("/heartbit", false)
				view.AddMessage([]byte(fmt.Sprintf("[blue]%s\n[orange]%s [white]%s", dataObj.Recipient, dataObj.AKA, dataObj.Text)))
			default:
				messageText := fmt.Sprintf(
					"[blue]Recipient:%s \n[green]Topics:%s [yellow]Seqno:%d\n[orange]%s:[white]%s",
					dataObj.Recipient, strings.Join(topicIDs, ";"), seqno, dataObj.AKA, dataObj.Text)
				view.AddMessage([]byte(messageText))
				//通知，有点吵，暂时关闭
				_ = beeep.Notify(dataObj.AKA, "Say:****", "")
			}

		}
	}

}

func infoAndHeartBitHandler(ctx context.Context, topic string) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):

			pubHandler("/heartbit", false)

			peers, err := ipfsapi.SubPeers(topic)
			if err != nil {
				continue
			}
			peersCount, ok := peers["Strings"].([]interface{})
			view.SetInfoView(fmt.Sprintf("[black]Topic:%s, AKA:%s, Peers:%d, ok?:%t", currentTopic, currentAKA, len(peersCount), ok))

			recs := secret.GetRecipients()

			view.SetRecipientListOptions(recs)
		}
	}
}

type Data struct {
	Text      string `json:"text"`
	AKA       string `json:"aka"`
	PubKey    string `json:"pubkey"`
	Recipient string `json:"recipient"`
	Encrypted bool   `json:"encrypted"`
}

//{"from":"12D3KooWFKQ8jcYyyDo245tFUnSZmMi2yZWmMaHdcUWWGqEYXqSZ","data":"ubmFtZSBwdWJzdWIgdGVzdCAyCg","seqno":"uFxhdm2UVf2Y","topicIDs":["uaGVsbG9kYXduZ3Jw"]}
type Message struct {
	From     string   `json:"from"`
	Data     string   `json:"data"`
	Seqno    string   `json:"seqno"`
	TopicIDs []string `json:"topicIDs"`
}
