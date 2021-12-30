package protocol

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
)

const (
	SIMPLE_STRING = '+'
	ERROR         = '-'
	INTEGER       = ':'
	BULK_STRINGS  = '$'
	ARRAYS        = '*'
)

var (
	errorMessageInvalidInput  = errors.New("failed to encode, invalid input")
	errorMessageInvalidSyntax = errors.New("failed to decode, invalid syntax")
)

func Encode(content interface{}) string {

	switch content.(type) {

	case int:
		return fmt.Sprintf("%c%d\r\n", INTEGER, content)

	case string:
		return fmt.Sprintf("%c%s\r\n", SIMPLE_STRING, content)

	case error:
		return fmt.Sprintf("%c%s\r\n", ERROR, content)

	case []byte:
		return fmt.Sprintf("%c%d\r\n%s\r\n", BULK_STRINGS, reflect.ValueOf(content).Len(), content)

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
	case BULK_STRINGS:
		return parseBulkString(content)
	case ARRAYS:
		return parseArray(content)
	}

	return errorMessageInvalidSyntax
}

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
	for start < len(input) && end < len(input) {

		if input[end] == '\r' && end+1 < len(input) && input[end+1] == '\n' {
			arr = append(arr, Decode(input[start:end]))
			end += 2
			start = end
			continue
		}
		end++
	}
	return arr
}
