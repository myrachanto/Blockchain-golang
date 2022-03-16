package main

import (
	"flag"
	"log"
)

func init() {
	log.SetPrefix("Blockchain: ")
}

func main() {
	port := flag.Uint("port", 5000, "TCP port number for BlockChain Server")
	flag.Parse()
	// fmt.Println(port)
	app := NewBlockchainServer(uint16(*port))
	app.Run()
}
