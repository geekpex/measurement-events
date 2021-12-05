package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"unsafe"
)

var (
	ErrIgnoreValue = errors.New("ignore value")
)

// String converts byte slice to string.
// --> Improves []byte to string conversion
func String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// ValueTime contains Time and Value integer
type ValueTime struct {
	Time  int
	Value int
}

func ParseValueTime(p []byte) (v ValueTime, err error) {
	var j int

loop:
	for i := 0; i < len(p); i++ {
		switch p[i] {
		case ',':
			v.Time, err = strconv.Atoi(String(p[:i]))
			if err != nil {
				// time is not integer
				err = ErrIgnoreValue
				return
			}
			j = i + 1
		case '\n', '\r':
			v.Value, err = strconv.Atoi(String(p[j:i]))
			if err != nil {
				// value is not integer
				err = ErrIgnoreValue
				return
			}
			if v.Value < -2 || v.Value > 2 {
				// invalid value received
				err = fmt.Errorf("value must be a integer greater than -3 and less than 3")
				return
			}
			break loop
		}
	}

	return
}

// measurementReader reads from input and sends parsed ValueTime values to output
func measurementReader(ctx context.Context, input io.Reader, exitWhenDone bool, output chan<- ValueTime) error {
	var readBuf = make([]byte, 32)
	var buf = make([]byte, 0, 32)
	var measurement ValueTime

	for {
		// exit if ctx has been cancelled
		if ctx.Err() != nil {
			return nil
		}

		n, err := input.Read(readBuf)
		if err == io.EOF {
			return nil
		} else if err != nil {
			panic("Failed to read from stdin: " + err.Error())
		}

		if n == 0 {
			continue
		}

		for i := 0; i < n; i++ {
			buf = append(buf, readBuf[i])
			if readBuf[i] != '\n' {
				continue
			}

			measurement, err = ParseValueTime(buf)
			if err != nil && err != ErrIgnoreValue {
				fmt.Fprintf(os.Stderr, "Received invalid input: %s\n", err.Error())
				buf = buf[:0]
				continue
			}

			output <- measurement
			buf = buf[:0]
		}

		if exitWhenDone && n < len(readBuf) {
			return io.EOF
		}

	}
}
