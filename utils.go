package main

import (
	"log"
	"runtime"
)

type Println func(...interface{})
type Printf func(string, ...interface{})

func checkError(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		log.Fatalf("%s:%d %v", file, line, err)
	}
}

func toMap(data interface{}) map[string]interface{} {
	if v, ok := data.(map[string]interface{}); ok {
		return v
	} else {
		log.Fatalf("convert: %v failed\n", v)
	}

	return nil
}
