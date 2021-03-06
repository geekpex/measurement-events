package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
)

var (
	usageText = `measurement-events

  -k  --keep-running	Exit only when interrupt or terminate is received
  -n  --no-headers	Do not ouput headers to STDOUT

This tool reads measurement info from stdin and writes warning events to stdout.
Supported input format:
Time,Level
1,0
2,1
4,1
5,0

Output format:
Start Time,End Time,Level
2,4,1
`
)

func main() {
	var exitWhenDone = true
	var noHeaders bool

	// Not that ideal solution, but in windows Ctrl+C also closes STDIN, but that does not happen in linux.
	// So in windows, the program waits for "clean" exist and in linux the STDIN reading is closed by exiting program.
	// If clean exit is wanted in linux, the Ctrl+D can be used (Sends EOF).
	var inWindows = runtime.GOOS == "windows"

	for i := 1; i < len(os.Args); i++ {
		switch strings.ToLower(os.Args[i]) {
		case "-k", "--keep-running":
			exitWhenDone = false
		case "-n", "--no-headers":
			noHeaders = true
		default:
			fmt.Print(usageText)
			return
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	var ch = make(chan ValueTime)
	var pub = EventPublisher{
		Output:    os.Stdout,
		NoHeaders: noHeaders,
	}

	var wg sync.WaitGroup
	if inWindows {
		wg.Add(2)
	} else {
		wg.Add(1)
	}

	// Start measurements reader
	go func() {
		err := measurementReader(ctx, os.Stdin, exitWhenDone, ch)
		if err == io.EOF {
			cancel()
		}

		if inWindows {
			wg.Done()
		}
	}()

	// Start process measurements
	go func() {
		pub.ProcessMeasurements(ctx, ch)
		wg.Done()
	}()

	// Exit safely if Interrupt or SIGTERM is received
	var exit = make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ctx.Done():
	case <-exit:
		cancel()
		// just a usablity thing when exiting using Ctrl+C
		fmt.Println()
	}
	// Wait for all pending work to finish
	wg.Wait()
}
