package main

import (
	"bufio"
	"log"
	"net"
)

func main() {
	client, err := net.Dial("tcp", "127.0.0.1:8001")
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(client)
	writer := bufio.NewWriter(client)
	nn, err := writer.Write([]byte("hi,world!\n"))
	if err != nil {
		panic(err)
	}
	writer.Flush()
	log.Println(nn)
	ml, _ := reader.ReadBytes('\n')
	log.Println(string(ml))
}
