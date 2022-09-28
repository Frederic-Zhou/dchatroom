package test

import (
	"fmt"
	"sync"
	"testing"
)

func TestSyncMap(t *testing.T) {
	syncMap := sync.Map{}

	syncMap.Store("a", "1")

	a, l := syncMap.LoadOrStore("a", "1")
	syncMap.LoadOrStore("b", "2")
	syncMap.LoadOrStore("c", "3")
	syncMap.LoadOrStore("d", "4")
	syncMap.LoadOrStore("e", "5")

	fmt.Println(a, l)
	a, l = syncMap.Load("a")
	fmt.Println(a, l)

	syncMap.Range(func(k, v interface{}) bool {
		fmt.Println(k, v)
		v = 2
		return true
	})

	a, l = syncMap.Load("a")
	fmt.Println(a, l)
}
