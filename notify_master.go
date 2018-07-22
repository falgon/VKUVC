package main

import (
	"./vip"
	"flag"
)

func init() {
	flag.Parse()
}

func main() {
	vip.SetRouteWithArgCheck()
}
