package protocol

import (
	"bytes"
	"errors"
	"log"
	"reflect"
	"testing"
)

func TestInt(t *testing.T) {

	testCases := []struct {
		input  int
		output string
	}{
		{1000, ":1000\r\n"},
		{-1, ":-1\r\n"},
		{0, ":0\r\n"},
	}

	for _, testCase := range testCases {

		// encode
		got := Encode(testCase.input)
		if got != testCase.output {
			log.Fatalf("encoding, input: %d, expected: %s, got: %s", testCase.input, testCase.output, got)
		}

		// decode
		decoded := Decode([]byte(testCase.output))
		if decoded != testCase.input {
			log.Fatalf("decoding, input: %s, expected: %d, got: %d", testCase.output, testCase.input, decoded)
		}
	}
}

func TestBulkString(t *testing.T) {
	testCases := []struct {
		input  []byte
		output string
	}{
		{[]byte("foobar"), "$6\r\nfoobar\r\n"},
		{[]byte(""), "$0\r\n\r\n"},
	}

	for _, testCase := range testCases {

		// encode
		got := Encode(testCase.input)
		if got != testCase.output {
			log.Fatalf("encoding, input: %v, expected: %s, got: %s", testCase.input, testCase.output, got)
		}

		// decode
		decoded := Decode([]byte(testCase.output))
		if !bytes.Equal(reflect.ValueOf(decoded).Bytes(), testCase.input) {
			log.Fatalf("decoding, input: %s, expected: %v, got: %v", testCase.output, testCase.input, decoded)
		}
	}
}

func TestString(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{"foobar", "+foobar\r\n"},
		{"", "+\r\n"},
	}

	for _, testCase := range testCases {

		// encode
		got := Encode(testCase.input)
		if got != testCase.output {
			log.Fatalf("encoding, input: %s, expected: %s, got: %s", testCase.input, testCase.output, got)
		}

		// decode
		decoded := Decode([]byte(testCase.output))
		if decoded != testCase.input {
			log.Fatalf("decoding, input: %s, expected: %s, got: %s", testCase.output, testCase.input, decoded)
		}
	}
}

func TestError(t *testing.T) {
	testCases := []struct {
		input  error
		output string
	}{
		{errors.New("foobar"), "-foobar\r\n"},
		{errors.New(""), "-\r\n"},
	}

	for _, testCase := range testCases {

		// encode
		got := Encode(testCase.input)
		if got != testCase.output {
			log.Fatalf("encoding, input: %s, expected: %s, got: %s", testCase.input, testCase.output, got)
		}

		// decode
		decoded := Decode([]byte(testCase.output))
		if !reflect.DeepEqual(decoded, testCase.input) {
			log.Fatalf("decoding, input: %s, expected: %s, got: %s", testCase.output, testCase.input, decoded)
		}
	}
}

func TestArray(t *testing.T) {
	testCases := []struct {
		input  []interface{}
		output string
	}{
		{[]interface{}{}, "*0\r\n"},
		{[]interface{}{"foo", "bar"}, "*2\r\n+foo\r\n+bar\r\n"},
		{[]interface{}{1, 2, 3}, "*3\r\n:1\r\n:2\r\n:3\r\n"},
	}

	for _, testCase := range testCases {

		// encode
		got := Encode(testCase.input)
		if got != testCase.output {
			log.Fatalf("encoding, input: %s, expected: %s, got: %s", testCase.input, testCase.output, got)
		}

		// decode
		decoded := Decode([]byte(testCase.output))
		if !reflect.DeepEqual(decoded, testCase.input) {
			log.Fatalf("decoding, input: %s, expected: %v, got: %v", testCase.output, testCase.input, decoded)
		}
	}
}
