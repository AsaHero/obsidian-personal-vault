package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	var address string
	flag.StringVar(&address, "address", "", "pass the address of the server")
	flag.Parse()

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("failed to dial to server: %v", err.Error())
	}
	defer conn.Close()

	fmt.Println("Connected to server. Type your messages below:")

	inputReader := bufio.NewReader(os.Stdin)
	scoketReader := bufio.NewReader(conn)
	for {
		fmt.Print("> ")
		request, err := inputReader.ReadString('\n')
		if err != nil {
			log.Printf("terminal input error: %v", err)
			return
		}

		_, err = conn.Write([]byte(request))
		if err != nil {
			log.Printf("Failed to send data: %v", err)
			return // Exit if the network connection breaks
		}

		response, err := scoketReader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\n[Server disconnected]")
			} else {
				fmt.Printf("\n[Read error: %v]\n", err)
			}
			return
		}

		fmt.Println(response)
	}
}
