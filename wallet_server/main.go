package main

import (
	"flag"
	"log"
)

func init() {
	log.SetPrefix("Wallet Server: ")
}
func main() {
	port := flag.Uint("port", 8080, "Tcp Port Number for wallet Server")
	gateway := flag.String("gateway", "http://127.0.0.1:5000", "Blockchain gateway")
	flag.Parse()
	log.Println(">>>>>>>>>>")

	app := NewWalletServer(uint16(*port), *gateway)
	app.Run()
}
