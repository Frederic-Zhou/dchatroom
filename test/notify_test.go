package test

import (
	"testing"

	"github.com/gen2brain/beeep"
)

func TestNotify(t *testing.T) {
	// err := beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
	// if err != nil {
	// 	panic(err)
	// }

	err := beeep.Notify("Title1", "Message body", "assets/information.png")
	if err != nil {
		panic(err)
	}

	err = beeep.Alert("Title2", "Message body", "assets/warning.png")
	if err != nil {
		panic(err)
	}
}
