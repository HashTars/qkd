package main

import (
	"log"
	"qkd/pkg"
)

func main() {
	pkg.Run(":8080")
}

func init() {
	log.SetPrefix("[QianKunDai] ")
}
