package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/viveknathani/retain/protocol"
)

func main() {

	host := flag.String("host", "127.0.0.1", "host")
	port := flag.Int("port", 8000, "port")
	flag.Parse()

	connection, err := net.Dial("tcp", *host+":"+fmt.Sprint(*port))
	handleError("client main: ", err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go waitForSignal(connection, sig, done)

	go func() {

		for {

			fmt.Print("> ")
			userReader := bufio.NewReader(os.Stdin)
			userInput, err := userReader.ReadBytes('\n')
			handleError("client main: ", err)

			arr := customSplit(userInput[0 : len(userInput)-1])
			_, err = connection.Write([]byte(protocol.Encode(arr)))
			handleError("client main: ", err)

			buffer := make([]byte, 65535)
			bytesRead, err := connection.Read(buffer)
			handleError("client main: ", err)
			buffer = buffer[0:bytesRead]
			decoded := protocol.Decode(buffer)

			switch decoded.(type) {
			case []interface{}:
				list := reflect.ValueOf(decoded)
				for i := 0; i < list.Len(); i++ {
					fmt.Printf(">>(%d) %s\n", i+1, list.Index(i))
				}
			default:
				fmt.Printf(">> %s\n", decoded)
			}
		}
	}()

	<-done
	fmt.Println("goodbye!")
}

func waitForSignal(connection net.Conn, sig <-chan os.Signal, done chan<- bool) {

	captured := <-sig
	arr := make([][]byte, 0)
	arr = append(arr, []byte("SAVE"))
	_, err := connection.Write([]byte(protocol.Encode(arr)))
	handleError("waitForSignal, upon saving:", err)
	fmt.Println()
	fmt.Println(captured)
	done <- true
}

func handleError(text string, err error) {

	if err != nil {
		log.Fatal(text, err)
	}
}

func customSplit(userInput []byte) [][]byte {

	result := make([][]byte, 0)

	i := 0
	// get the first segment, the command
	temp := make([]byte, 0)
	for ; i < len(userInput); i++ {

		if userInput[i] == ' ' {
			i++
			break
		}

		temp = append(temp, userInput[i])
	}
	result = append(result, temp)

	captureUntilNextQuote := false

	temp = make([]byte, 0)
	for ; i < len(userInput); i++ {

		if userInput[i] == '"' {
			captureUntilNextQuote = !captureUntilNextQuote
			continue
		}

		if userInput[i] == ' ' && !captureUntilNextQuote {
			result = append(result, temp)
			temp = make([]byte, 0)
			continue
		}

		temp = append(temp, userInput[i])
	}

	if len(temp) != 0 {
		result = append(result, temp)
	}
	return result
}
