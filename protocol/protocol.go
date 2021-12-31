// this package implements the RESP protocol (https://redis.io/topics/protocol)
package protocol

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
)

// first byte for corresponding type of data
const (
	SIMPLE_STRING = '+'
	ERROR         = '-'
	INTEGER       = ':'
	DOUBLE        = ','
	BULK_STRINGS  = '$'
	ARRAYS        = '*'
)

var (
	errorMessageInvalidInput  = errors.New("failed to encode, invalid input")
	errorMessageInvalidSyntax = errors.New("failed to decode, invalid syntax")
)

// Encode will take variable type of input as content and
// output a RESP-compliant string
func Encode(content interface{}) string {

	switch content.(type) {

	case int:
		return fmt.Sprintf("%c%d\r\n", INTEGER, content)

	case float64:
		return fmt.Sprintf("%c%f\r\n", DOUBLE, content)

	case string:
		return fmt.Sprintf("%c%s\r\n", SIMPLE_STRING, content)

	case error:
		return fmt.Sprintf("%c%s\r\n", ERROR, content)

	case []byte:
		return fmt.Sprintf("%c%d\r\n%s\r\n", BULK_STRINGS, reflect.ValueOf(content).Len(), content)

	case [][]byte:
		arr := reflect.ValueOf(content)
		res := fmt.Sprintf("%c%d\r\n", ARRAYS, arr.Len())

		for i := 0; i < arr.Len(); i++ {
			res += Encode(arr.Index(i).Interface().([]byte))
		}
		return res

	case []interface{}:
		arr := reflect.ValueOf(content)
		res := fmt.Sprintf("%c%d\r\n", ARRAYS, arr.Len())

		for i := 0; i < arr.Len(); i++ {
			res += Encode(arr.Index(i).Interface())
		}
		return res
	}

	panic(errorMessageInvalidInput)
}

// Decode expects a RESP-compliant input and outputs the
// corresponding data
func Decode(content []byte) interface{} {

	if len(content) == 0 {
		return errorMessageInvalidSyntax
	}

	switch content[0] {

	case SIMPLE_STRING:
		return parseSimpleString(content)
	case ERROR:
		return parseError(content)
	case INTEGER:
		return parseInteger(content)
	case DOUBLE:
		return parseDouble(content)
	case BULK_STRINGS:
		return parseBulkString(content)
	case ARRAYS:
		return parseArray(content)
	}

	return errorMessageInvalidSyntax
}

// excludeCRLFAndReturnIndex will read data until the first
// occurence of CRLF
func excludeCRLFAndReturnIndex(input []byte) ([]byte, int) {

	res := make([]byte, 0)
	lastValidIndex := -1
	for i := 0; i < len(input); i++ {

		if input[i] == '\r' && i+1 < len(input) && input[i+1] == '\n' {
			lastValidIndex = i
			break
		}

		res = append(res, input[i])
	}

	return res, lastValidIndex
}

func handleError(text string, err error) {

	if err != nil {
		log.Fatal(text, err)
	}
}

func parseSimpleString(input []byte) string {

	output, _ := excludeCRLFAndReturnIndex(input[1:])
	return string(output)
}

func parseError(input []byte) error {

	output, _ := excludeCRLFAndReturnIndex(input[1:])
	return errors.New(string(output))
}

func parseInteger(input []byte) int {

	output, _ := excludeCRLFAndReturnIndex(input[1:])
	num, err := strconv.Atoi(string(output))
	handleError("parseError: ", err)
	return num
}

func parseDouble(input []byte) float64 {

	output, _ := excludeCRLFAndReturnIndex(input[1:])
	num, err := strconv.ParseFloat(string(output), 64)
	handleError("parseDouble: ", err)
	return num
}

func parseBulkString(input []byte) []byte {

	_, idx := excludeCRLFAndReturnIndex(input[1:])
	output, _ := excludeCRLFAndReturnIndex(input[idx+3:]) // skip \r\n and 1 more since it is based on input[1:]
	return output
}

func parseArray(input []byte) []interface{} {

	temp, idx := excludeCRLFAndReturnIndex(input[1:])
	_, err := strconv.Atoi(string(temp))
	handleError("parseArray: ", err)

	arr := make([]interface{}, 0)

	start := idx + 3
	end := start

	ignoreFirstInstanceCRLF := false
	for start < len(input) && end < len(input) {

		if input[end] == '$' {
			ignoreFirstInstanceCRLF = true
		}

		if input[end] == '\r' && end+1 < len(input) && input[end+1] == '\n' {

			if ignoreFirstInstanceCRLF {
				end++
				ignoreFirstInstanceCRLF = false
				continue
			}
			arr = append(arr, Decode(input[start:end]))
			end += 2
			start = end
			continue
		}
		end++
	}
	return arr
}
