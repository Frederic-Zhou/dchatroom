package main

import (
	"context"
	"fmt"
	"myd/ipfsapi"
	"myd/view"
	"strings"
)

var currentTopic string = ""
var currentSubChan chan []byte
var cancelSub context.CancelFunc

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

	view.AddMessage([]byte(currentTopic + "/" + text))
	switch {
	case strings.HasPrefix(text, "/sub "):

		topic := strings.TrimLeft(text, "/sub")
		topic = strings.TrimSpace(topic)

		if cancelSub != nil {
			cancelSub()
		}

		var ctx context.Context
		ctx, cancelSub = context.WithCancel(context.Background())

		var err error
		currentSubChan, err = ipfsapi.Sub(ctx, topic)
		if err != nil {
			view.SetInfoView("[red]sub error:[white]" + err.Error())
			return
		}

		view.AddMessage([]byte("sub success:" + topic))

		currentTopic = topic

		//todo:parse sub message
		//include decrypt
		go func() {
			for line := range currentSubChan {
				view.AddMessage(line)
			}
		}()

	case strings.HasPrefix(text, "/aka "):

		view.AddMessage([]byte("My name is:" + text))
	default:

		if currentTopic == "" {
			view.SetInfoView("[red]sub a topic first[white]")
			return
		}
		//todo: format text to a message struct
		//include encrypt
		_, err := ipfsapi.Pub(currentTopic, text)
		if err != nil {
			view.SetInfoView("[red]error:[white]" + err.Error())
		}
	}

}
