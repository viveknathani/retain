package main

import (
	"errors"
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
		response := executeCommand(store, arr.([]interface{}))
		_, err = connection.Write(response)
		handleError("serve: ", err)
	}
}

func executeCommand(store *store.Storage, respArray []interface{}) protocol.RespEncodedString {

	response := ""
	errorMessage := protocol.Encode(errors.New("invalid command syntax"))
	switch string(respArray[0].([]byte)) {

	case "PING":
		if len(respArray) > 1 {
			response = string(respArray[1].([]byte))
			break
		}
		response = "PONG"

	case "ECHO":
		if len(respArray) > 1 {
			response = string(respArray[1].([]byte))
		}

	case "SET":
		if len(respArray) != 3 {
			return errorMessage
		}

		key := respArray[1].([]byte)
		value := respArray[2].([]byte)
		store.Set(key, value)
		response = "OK"

	case "GET":
		if len(respArray) != 2 {
			return errorMessage
		}

		key := respArray[1].([]byte)
		value, ok := store.Get(key)

		if !ok {
			return protocol.Encode(errors.New("(nil)"))
		}
		response = string(value.([]byte))

	case "DEL":
		if len(respArray) != 2 {
			return errorMessage
		}

		key := respArray[1].([]byte)
		store.Delete(key)
		response = "OK"

	case "MSET":

		if len(respArray)%2 == 0 || len(respArray) == 1 {
			return errorMessage
		}

		for i := 1; i < len(respArray); i += 2 {
			key := respArray[i].([]byte)
			value := respArray[i+1].([]byte)
			store.Set(key, value)
		}

		response = "OK"

	case "MGET":

		if len(respArray) == 1 {
			return errorMessage
		}

		arr := make([][]byte, 0)
		for i := 1; i < len(respArray); i++ {
			key := respArray[i].([]byte)
			value, ok := store.Get(key)
			if !ok {
				arr = append(arr, []byte("(nil)"))
				continue
			}
			arr = append(arr, value.([]byte))
		}

		return protocol.Encode(arr)

	case "SAVE":
		if len(respArray) != 1 {
			return errorMessage
		}
		store.Save()

		response = "OK"
	default:
		return errorMessage
	}

	return protocol.Encode(response)
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
