package ipfsapi

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	mbase "github.com/multiformats/go-multibase"
)

var BaseUrl string

// type ApiResponse map[string]interface{}
type ApiResponse datamodel.Node

func init() {
	port := flag.Int("p", 5001, "ipfs port number, Default: 5001")
	BaseUrl = fmt.Sprintf("http://127.0.0.1:%d/api/v0/", *port)
}

func Version() (version ApiResponse, err error) {
	version, err = invokeApi("version", nil, nil, nil, nil, context.Background())
	if err != nil {
		return
	}
	return
}

func Sub(ctx context.Context, topic string) (subChan chan []byte, err error) {

	topic = Base64urlEncode([]byte(topic))

	args := url.Values{
		"arg": []string{topic},
	}

	subChan = make(chan []byte, 10)

	go func(c chan []byte, ctx context.Context) {
		_, err = invokeApi("pubsub/sub", args, nil, nil, c, ctx)
		if err != nil {
			return
		}
	}(subChan, ctx)

	return
}

func Pub(topic string, node datamodel.Node) (pub ApiResponse, err error) {

	topic = Base64urlEncode([]byte(topic))

	b := bytes.NewBuffer([]byte{})
	err = dagjson.Encode(node, b)
	if err != nil {
		return
	}

	args := url.Values{
		"arg": []string{topic},
	}

	// 实例化multipart
	body := &bytes.Buffer{}
	header := http.Header{}

	//refer:https://blog.csdn.net/huobo123/article/details/104288030
	//=========================================================
	writer := multipart.NewWriter(body)

	// 创建multipart 文件字段
	part, err := writer.CreateFormField("file")
	if err != nil {
		return nil, err
	}
	// 写入文件数据到multipart，和读取本地文件方法的唯一区别
	_, err = part.Write(b.Bytes())
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	header.Add("Content-Type", writer.FormDataContentType())

	pub, err = invokeApi("pubsub/pub", args, header, body, nil, context.Background())

	return
}

func SubLs() (ls ApiResponse, err error) {
	ls, err = invokeApi("pubsub/ls", nil, nil, nil, nil, context.Background())
	if err != nil {
		// fmt.Println(err)
		return
	}

	return
}

func SubPeers(topic string) (peers ApiResponse, err error) {
	encoder, _ := mbase.EncoderByName("base64url")
	topic = encoder.Encode([]byte(topic))

	args := url.Values{
		"arg": []string{topic},
	}

	peers, err = invokeApi("pubsub/peers", args, nil, nil, nil, context.Background())
	if err != nil {
		return
	}

	return
}

func DagPut(node datamodel.Node) (putResult ApiResponse, err error) {

	// 实例化multipart
	body := &bytes.Buffer{}
	header := http.Header{}

	b := bytes.NewBuffer([]byte{})
	err = dagjson.Encode(node, b)
	if err != nil {
		return
	}

	//refer:https://blog.csdn.net/huobo123/article/details/104288030
	//=========================================================
	writer := multipart.NewWriter(body)

	// 创建multipart 文件字段
	part, err := writer.CreateFormField("file")
	if err != nil {
		return nil, err
	}
	// 写入文件数据到multipart，和读取本地文件方法的唯一区别
	_, err = part.Write(b.Bytes())
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	header.Add("Content-Type", writer.FormDataContentType())

	putResult, err = invokeApi("dag/put", nil, header, body, nil, context.Background())

	return
}

func invokeApi(path string, args url.Values, header http.Header, body io.Reader, c chan []byte, ctx context.Context) (apiResponse ApiResponse, err error) {

	// fmt.Println("api:", fmt.Sprintf("%s%s?%s", BaseUrl, path, args.Encode()))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s?%s", BaseUrl, path, args.Encode()), body)
	if err != nil {
		fmt.Println("new req err:", err)
		return
	}

	req.Header = header

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// fmt.Println("request err:", err)
		return
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	respBody := []byte{}

	streamChan := make(chan []byte, 10)
	bodyReadCtx, cancel := context.WithCancel(ctx)

	go func(streamChan chan []byte) {
		for {
			line, err := reader.ReadBytes('\n')
			// fmt.Println("read line", err)
			if err != nil {
				cancel()
				return
			}
			// fmt.Println(string(line))
			streamChan <- line
		}
	}(streamChan)

ReadLoop:
	for {
		select {
		case <-bodyReadCtx.Done():
			break ReadLoop
		case line := <-streamChan:
			if c != nil {
				c <- line
			} else {
				respBody = append(respBody, line...)
			}
		}
	}

	if len(respBody) > 0 {

		serial := bytes.NewReader(respBody)

		// np := basicnode.Prototype.Any              // Pick a stle for the in-memory data.
		nb := basicnode.Prototype.Any.NewBuilder() // Create a builder.
		dagjson.Decode(nb, serial)                 // Hand the builder to decoding -- decoding will fill it in!
		apiResponse = nb.Build()                   // Call 'Build' to get the resulting Node.  (It's immutable!)

	}

	return

}

func Base64urlEncode(data []byte) string {
	encoder, _ := mbase.EncoderByName("base64url")
	return encoder.Encode(data)
}
func Base64urlDecode(data string) []byte {
	_, result, _ := mbase.Decode(data)
	return result
}
