package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

var (
	usageText = `measurement-events

	-k	--keep-running	Exit only when interrupt or terminate is received

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

	for i := 1; i < len(os.Args); i++ {
		switch strings.ToLower(os.Args[i]) {
		case "-k", "--keep-running":
			exitWhenDone = false
		default:
			fmt.Print(usageText)
			return
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	var ch = make(chan ValueTime)
	var pub = EventPublisher{
		Output: os.Stdout,
	}

	var wg sync.WaitGroup

	// Start measurements reader
	go func() {
		err := measurementReader(ctx, os.Stdin, exitWhenDone, ch)
		if err == io.EOF {
			cancel()
		}
	}()

	// Start process measurements
	wg.Add(1)
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
