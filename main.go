package main

import (
	"context"
	"encoding/json"
	"fmt"
	"myd/ipfsapi"
	"myd/secret"
	"myd/view"
	"os"
	"strings"
	"time"
)

var currentTopic string = ""
var currentSubChan chan []byte
var cancelSub context.CancelFunc
var currentAKA, _ = os.Hostname()

func main() {

	// ipfsapi.Version()

	// ipfsapi.SubLs()
	// ipfsapi.SubPeers("hellodawngrp")
	// ipfsapi.Pub("hellodawngrp", "hello")
	// c, _ := ipfsapi.Sub("hellodawngrp")
	// for line := range c {
	// 	fmt.Print(string(line))
	// }

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
		topic := strings.TrimSpace(strings.TrimPrefix(text, "/sub"))
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
		go infoHandler(ctx, topic)
		//开goroutin不断获取ipfsapi sub的通道数据
		go messageHandler(ctx, currentSubChan)

	case strings.HasPrefix(text, "/aka "):
		aka := strings.TrimSpace(strings.TrimPrefix(text, "/aka"))
		if aka == "" {
			return
		}
		currentAKA = aka
		view.AddMessage([]byte("My name is:" + text))
	default: // default is pub messag to topic

		if currentTopic == "" {
			view.SetInfoView("[red]sub a topic first[white]")
			return
		}

		if text == "" {
			return
		}

		text, err := secret.Encrypt(text)
		if err != nil {
			view.AddMessage([]byte("Encrypt Error:" + err.Error()))
			return
		}
		data, _ := json.Marshal(Data{
			Text:      text,
			AKA:       currentAKA,
			PubKey:    "",
			Recipient: secret.GetLocalRecipient(),
		})

		//todo: signtrue
		_, err = ipfsapi.Pub(currentTopic, string(data))
		if err != nil {
			view.SetInfoView("[red]error:[white]" + err.Error())
		}
	}

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
			data := ipfsapi.Base64urlDecode(message.Data)
			dataObj := Data{}

			if err := json.Unmarshal(data, &dataObj); err != nil {
				view.AddMessage([]byte("json2:" + err.Error() + "\n" + string(data)))
				return
			}

			recipients := secret.AddRemoteRecipient(dataObj.Recipient)

			topicIDs := []string{}
			for _, t := range message.TopicIDs {
				topicIDs = append(topicIDs, string(ipfsapi.Base64urlDecode(t)))
			}

			//todo: check signtrue

			planText, err := secret.Decrypt(dataObj.Text)
			if err != nil {
				planText = "[red]Decrypt Error:" + err.Error()
			}

			messageText := fmt.Sprintf(
				"[blue]PeerID:%s [green]Topics:%s\n[orange]%s:[white]%s\n[gray] (r=%d) %s",
				message.From, strings.Join(topicIDs, ";"), dataObj.AKA, planText, len(recipients), dataObj.Text)

			view.AddMessage([]byte(messageText))
			//通知，有点吵，暂时关闭
			// _ = beeep.Notify("New Message", planText, "")

		}
	}

}

func infoHandler(ctx context.Context, topic string) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
			peers, _ := ipfsapi.SubPeers(topic)
			peersCount := peers["Strings"].([]interface{})

			view.SetInfoView(fmt.Sprintf("[black]Topic:%s, AKA:%s, Peers:%d", currentTopic, currentAKA, len(peersCount)))
		}
	}
}

type Data struct {
	Text      string `json:"text"`
	AKA       string `json:"aka"`
	PubKey    string `json:"pubkey"`
	Recipient string `json:"recipient"`
}

//{"from":"12D3KooWFKQ8jcYyyDo245tFUnSZmMi2yZWmMaHdcUWWGqEYXqSZ","data":"ubmFtZSBwdWJzdWIgdGVzdCAyCg","seqno":"uFxhdm2UVf2Y","topicIDs":["uaGVsbG9kYXduZ3Jw"]}
type Message struct {
	From     string   `json:"from"`
	Data     string   `json:"data"`
	Seqno    string   `json:"seqno"`
	TopicIDs []string `json:"topicIDs"`
}
