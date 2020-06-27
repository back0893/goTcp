package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
)

func main() {
	client1, err := net.Dial("tcp", "127.0.0.1:10086")
	if err != nil {
		panic(err)
	}
	client2, err := net.Dial("tcp", "127.0.0.1:10086")
	if err != nil {
		panic(err)
	}
	client3, err := net.Dial("tcp", "127.0.0.1:10086")
	if err != nil {
		panic(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(3)
	for index, con := range []net.Conn{client1, client2, client3} {
		go func(i int, con net.Conn) {
			reader := bufio.NewReader(con)
			nn, err := con.Write([]byte(fmt.Sprintf("%d...hi,world!\n", i)))
			if err != nil {
				panic(err)
			}
			log.Println(nn)
			ml := make([]byte, 1024)
			n, _ := reader.Read(ml)
			log.Println(string(ml[:n]))
			wg.Done()
		}(index+1, con)
	}
	wg.Wait()
	client1.Close()
	client2.Close()
	client3.Close()

}
