package store

import (
	"log"
	"os"
	"testing"
)

var testCases = []struct {
	key   RetainKey
	value interface{}
}{
	{key: RetainKey("hello"), value: 5},
	{key: RetainKey("hey"), value: "56"},
}

func TestGetAndSet(t *testing.T) {

	mp := New()
	for _, testCase := range testCases {

		mp.Set(testCase.key, testCase.value)

		v, ok := mp.Get(testCase.key)

		if !ok || v != testCase.value {
			log.Fatalf("failed Get at %v, got: %v", testCase, v)
		}
	}
}

func TestDelete(t *testing.T) {

	mp := New()
	for _, testCase := range testCases {

		mp.Set(testCase.key, testCase.value)
		mp.Delete(testCase.key)
		_, ok := mp.Get(testCase.key)

		if ok {
			log.Fatalf("failed Delete at %v", testCase)
		}
	}
}

func TestSaveAndLoad(t *testing.T) {

	mp := New()
	for _, testCase := range testCases {
		mp.Set(testCase.key, testCase.value)
	}

	mp.Save()

	for _, testCase := range testCases {
		mp.Delete(testCase.key)
	}

	mp.LoadFromDisk(fileName)

	for _, testCase := range testCases {

		v, ok := mp.Get(testCase.key)
		if !ok || v != testCase.value {
			log.Fatalf("failed TestLoadAndSave at %v", testCase)
		}
	}

	// clean up
	err := os.Remove(fileName)
	handleError("failed TestLoadAndSave cleanup", err)
}
