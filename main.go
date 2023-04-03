package main

import (
	"github.com/billikeu/chatgpt-server/backend"
	"github.com/billikeu/chatgpt-server/backend/conf"
)

func main() {
	server := backend.NewServer(&conf.Config{
		Host:      "127.0.0.1",
		Port:      8089,
		Proxy:     "", // optional http://127.0.0.1:10809, socks5://127.0.0.1:10808
		SecretKey: "", // sk-dd3434545 your secret key
	})
	server.Start()
}
