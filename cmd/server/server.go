package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"syscall"

	"github.com/viveknathani/retain/protocol"
	"github.com/viveknathani/retain/store"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

func printColor(colorName string) {

	os := runtime.GOOS
	if os != "windows" {
		fmt.Print(colorName)
	}
}

func serve(store *store.Storage, connection net.Conn) {

	defer connection.Close()

	address := connection.RemoteAddr().String()
	printColor(colorGreen)
	fmt.Printf("new client => %s\n", address)
	printColor(colorReset)

	for {
		buffer := make([]byte, 65535)
		bytesRead, err := connection.Read(buffer)
		if handleErrorWhileServing(address, err) {
			break
		}

		buffer = buffer[0:bytesRead]
		arr := protocol.Decode(buffer)
		response := executeCommand(store, arr.([]interface{}), address)
		_, err = connection.Write(response)
		if handleErrorWhileServing(address, err) {
			break
		}
	}
}

func handleErrorWhileServing(address string, err error) bool {

	if err == nil {
		return false
	}

	printColor(colorRed)
	fmt.Printf("[%s] > ", address)
	if checkIfClientLeft(err) {
		fmt.Printf("(left)\n")
		printColor(colorReset)
		return true
	}

	if checkIfWeLostConnection(err) {
		fmt.Printf("(ECONNRESET)\n")
		printColor(colorReset)
		return true
	}

	fmt.Printf("%s\n", err.Error())
	printColor(colorReset)
	return true
}

func checkIfClientLeft(err error) bool {
	return errors.Is(err, io.EOF)
}

func checkIfWeLostConnection(err error) bool {
	return errors.Is(err, syscall.ECONNRESET)
}

func executeCommand(store *store.Storage, respArray []interface{}, address string) protocol.RespEncodedString {

	response := ""
	errorMessage := protocol.Encode(errors.New("invalid command syntax"))

	command := string(respArray[0].([]byte))
	printColor(colorYellow)
	fmt.Printf("[%s] > request for %s\n", address, command)
	printColor(colorReset)

	switch command {

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

	storage, loadedFromDisk := store.New()

	if loadedFromDisk {
		fmt.Println("loaded from disk (retain.db)")
	}

	for {
		connection, err := listener.Accept()
		handleError("server main: ", err)
		go serve(storage, connection)
	}
}

func handleError(text string, err error) {

	if err != nil {
		printColor(colorRed)
		fmt.Println(text, err)
		printColor(colorReset)
		os.Exit(1)
	}
}
