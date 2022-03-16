package main

import (
	"fmt"

	"github.com/myrachanto/blockchain/utils"
)

func main() {
	// fmt.Println(utils.IsFoundHost("127.0.0.1", 5000))
	fmt.Println(utils.FindNeighbors("127.0.0.1", 5000, 0, 3, 5000, 5003))
}
