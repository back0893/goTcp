package main

import (
	"bufio"
	"log"
	"net"
)

func main() {
	client, err := net.Dial("tcp", "127.0.0.1:10086")
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
	ml := make([]byte, 1024)
	n, _ := reader.Read(ml)
	log.Println(string(ml[:n]))
}
