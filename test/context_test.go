package test

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan string, 1)
	//c chan string, ctx context.Context
	go func() {

	forLoop:
		for {
			select {
			case <-ctx.Done():
				fmt.Println("done")
				break forLoop
			case txt := <-c:
				fmt.Println(txt)
			}
		}

	}()

	for i := 0; i < 5; i++ {
		c <- fmt.Sprintf("%d", i)

		time.Sleep(1 * time.Second)
		cancel()
		fmt.Println("call done")
	}

	select {}
}
