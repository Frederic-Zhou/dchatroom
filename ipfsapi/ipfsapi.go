package ipfsapi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	mbase "github.com/multiformats/go-multibase"
)

var BaseUrl string

type ApiResponse map[string]interface{}

func init() {
	port := flag.Int("p", 5001, "ipfs port number")
	flag.Parse()
	BaseUrl = fmt.Sprintf("http://127.0.0.1:%d/api/v0/", *port)
}

func Version() (version ApiResponse, err error) {
	ctx := context.Background()
	version, err = invokeApi("version", nil, nil, nil, nil, ctx)
	if err != nil {
		return
	}
	return
}

func Sub(topic string) (subChan chan []byte, err error) {

	encoder, _ := mbase.EncoderByName("base64url")
	topic = encoder.Encode([]byte(topic))

	args := url.Values{
		"arg": []string{topic},
	}

	subChan = make(chan []byte, 10)
	ctx := context.Background()

	go func(c chan []byte, ctx context.Context) {
		_, err = invokeApi("pubsub/sub", args, nil, nil, c, ctx)
		if err != nil {
			return
		}
	}(subChan, ctx)

	return
}

func Pub(topic string, content string) (pub ApiResponse, err error) {

	encoder, _ := mbase.EncoderByName("base64url")
	topic = encoder.Encode([]byte(topic))

	args := url.Values{
		"arg": []string{topic},
	}

	ctx := context.Background()

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
	_, err = part.Write([]byte(content))
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	header.Add("Content-Type", writer.FormDataContentType())

	pub, err = invokeApi("pubsub/pub", args, header, body, nil, ctx)
	if err != nil {
		return
	}
	return
}

func SubLs() (ls ApiResponse, err error) {
	ctx := context.Background()
	ls, err = invokeApi("pubsub/ls", nil, nil, nil, nil, ctx)
	if err != nil {
		fmt.Println(err)
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

	ctx := context.Background()

	peers, err = invokeApi("pubsub/peers", args, nil, nil, nil, ctx)
	if err != nil {
		return
	}

	return
}

func invokeApi(path string, args url.Values, header http.Header, body io.Reader, c chan []byte, ctx context.Context) (apiResponse ApiResponse, err error) {

	fmt.Println("api:", fmt.Sprintf("%s%s?%s", BaseUrl, path, args.Encode()))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s?%s", BaseUrl, path, args.Encode()), body)
	if err != nil {
		fmt.Println("new req err:", err)
		return
	}

	req.Header = header

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("request err:", err)
		return
	}

	reader := bufio.NewReader(resp.Body)
	respBody := []byte{}

ReadLoop:
	for {
		select {
		case <-ctx.Done():
			break ReadLoop
		default:
			line, err := reader.ReadBytes('\n')
			if err != nil {
				break ReadLoop
			}

			if c != nil {
				c <- line
			} else {
				respBody = append(respBody, line...)
			}
		}
	}

	if len(respBody) > 0 {
		err = json.Unmarshal(respBody, &apiResponse)
		if err != nil {
			fmt.Println("jsonUnmarshal error:", err)
		}
	}
	fmt.Println("respBody:", string(respBody), err)
	return

}
