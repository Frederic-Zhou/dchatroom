package main

import (
	"fmt"
	"myd/ipfsapi"
)

func main() {

	ipfsapi.Version()

	ipfsapi.SubLs()
	ipfsapi.SubPeers("hellodawngrp")
	ipfsapi.Pub("hellodawngrp", "hello")
	c, _ := ipfsapi.Sub("hellodawngrp")
	for line := range c {
		fmt.Print(string(line))
	}
}
