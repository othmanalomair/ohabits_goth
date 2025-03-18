package main

import (
	"ohabits.com/cmd/server"
	"ohabits.com/internal/db"
)

func main() {
	db.Connect()
	defer db.Close()

	server.Server()

}
