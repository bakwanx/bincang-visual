package main

import (
	"bincang-visual/server"
	"log"
)

func main() {
	err := server.Run()
	if err != nil {
		log.Fatal("server error, %s", err.Error())
	}
}
