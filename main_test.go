package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"path"
	"sync"
	"testing"
)

func BenchmarkProcessing(b *testing.B) {
	outputBuf.Grow(len(expectedOutput))

	input := bytes.NewReader(inputData)

	ev := NewEventPublisher(outputBuf)

	var ctx context.Context
	var cancel func()

	ch := make(chan ValueTime)
	var wg sync.WaitGroup

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		outputBuf.Reset()
		input.Reset(inputData)
		ev.Reset()

		ctx, cancel = context.WithCancel(context.Background())

		wg.Add(2)
		b.StartTimer()
		go func() {
			measurementReader(ctx, input, true, ch)
			cancel()
			wg.Done()
		}()

		go func() {
			for {
				select {
				case <-ctx.Done():
					wg.Done()
					return
				case <-ch:
				}
			}

			// ev.ProcessMeasurements(ctx, ch)
			// wg.Done()
		}()

		wg.Wait()

		// if !bytes.Equal(outputBuf.Bytes(), expectedOutput) {
		// 	b.Fatalf("Output does not match expected data:\n%s\n%s\n", outputBuf.String(), string(expectedOutput))
		// }
	}

}

func BenchmarkProcessingTenThousand(b *testing.B) {
	outputBuf.Grow(500 * 1024)

	input := bytes.NewReader(tenThousandInput)

	ev := NewEventPublisher(outputBuf)

	var ctx context.Context
	var cancel func()

	ch := make(chan ValueTime)
	var wg sync.WaitGroup

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		outputBuf.Reset()
		input.Reset(tenThousandInput)
		ev.Reset()

		ctx, cancel = context.WithCancel(context.Background())

		wg.Add(2)
		go func() {
			measurementReader(ctx, input, true, ch)
			cancel()
			wg.Done()
		}()

		go func() {
			ev.ProcessMeasurements(ctx, ch)
			wg.Done()
		}()

		wg.Wait()

		if !bytes.Equal(outputBuf.Bytes(), expectedTenThousand) {
			b.Fatalf("Output does not match expected data:\n%s\n%s\n", outputBuf.String(), string(expectedOutput))
		}
	}

}

var (
	inputData = []byte(`Time,Level
1,0
2,1
4,1
5,0
10,0
12,2
17,2
18,1
24,1
25,-2
26,-1
28,-1
29,0
30,1
32,2
33,2
34,1
36,1
37,2
38,2
40,0
`)
	expectedOutput = []byte(`Start Time,End Time,Level
2,4,1
12,24,1
12,17,2
25,28,-1
25,25,-2
30,38,1
32,33,2
37,38,2
`)
	outputBuf = &bytes.Buffer{}

	tenThousandInput, _    = ioutil.ReadFile(path.Join("testdata", "10000_rows.csv"))
	expectedTenThousand, _ = ioutil.ReadFile(path.Join("testdata", "expected_10000_rows.csv"))
)
