package util

import (
	"bytes"
	"log"
	"runtime"
	"strconv"
	"time"
)

func Trace(traceId string, msg string) func() {

	start := time.Now()
	log.Printf("%s,Enter %s.\n", traceId, msg)

	return func() {
		log.Printf("%s,Exit %s (%s)\n", traceId, msg, time.Since(start))
	}
}

func GetGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
