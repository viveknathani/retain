package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/viveknathani/retain/protocol"
	"github.com/viveknathani/retain/store"
)

func serve(store *store.Storage, connection net.Conn) {

	defer connection.Close()

	for {
		buffer := make([]byte, 65535)
		bytesRead, err := connection.Read(buffer)

		if err != nil {
			fmt.Println("serve:", err)
			return
		}

		buffer = buffer[0:bytesRead]
		arr := protocol.Decode(buffer)
		response := executeCommand(arr.([]interface{}))
		_, err = connection.Write([]byte(protocol.Encode(response)))
		handleError("serve: ", err)
	}
}

func executeCommand(respArray []interface{}) string {

	switch string(respArray[0].([]byte)) {

	case "PING":
		if len(respArray) > 1 {
			return string(respArray[1].([]byte))
		}
		return "PONG"

	case "ECHO":
		if len(respArray) > 1 {
			return string(respArray[1].([]byte))
		}
		return ""

	case "SET":
	case "GET":
	case "DEL":
	case "MSET":
	case "MGET":
	}

	return "nothing"
}

func main() {

	host := flag.String("host", "127.0.0.1", "host")
	port := flag.Int("port", 8000, "port")
	flag.Parse()

	listener, err := net.Listen("tcp", *host+":"+fmt.Sprint(*port))
	handleError("server main: ", err)

	storage := store.New()

	for {
		connection, err := listener.Accept()
		handleError("server main: ", err)
		go serve(storage, connection)
	}
}

func handleError(text string, err error) {

	if err != nil {
		log.Fatal(text, err)
	}
}
