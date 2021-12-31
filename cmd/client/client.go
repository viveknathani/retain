package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/viveknathani/retain/protocol"
)

func main() {

	host := flag.String("host", "127.0.0.1", "host")
	port := flag.Int("port", 8000, "port")
	flag.Parse()

	connection, err := net.Dial("tcp", *host+":"+fmt.Sprint(*port))
	handleError("client main: ", err)

	for {

		fmt.Print("> ")
		userReader := bufio.NewReader(os.Stdin)
		userInput, err := userReader.ReadBytes('\n')
		handleError("client main: ", err)

		arr := bytes.Split(userInput[0:len(userInput)-1], []byte(" "))
		_, err = connection.Write([]byte(protocol.Encode(arr)))
		handleError("client main: ", err)

		buffer := make([]byte, 65535)
		bytesRead, err := connection.Read(buffer)
		handleError("client main: ", err)
		buffer = buffer[0:bytesRead]
		fmt.Printf(">> %s\n", protocol.Decode(buffer))
	}
}

func handleError(text string, err error) {

	if err != nil {
		log.Fatal(text, err)
	}
}
