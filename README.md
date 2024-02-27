# Batching

This batching library could be used in processing pipelines where individual tasks are grouped together into small batches. This can improve throughput by reducing the number of requests made to a downstream system.

```bash
go get github.com/xsoroton/batching
```

### Test

```bash
go test -v -race
```

## Batching Library Example

```bash
go run ./example
```

[main.go](example%2Fmain.go)

```go
package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/xsoroton/batching"
)

/*
======================================================
             BATCHING LIBRARY USE EXAMPLE
======================================================
*/

type processor struct{}

// example of local implementation for BatchProcessor
func (b *processor) BatchProcessor(jobs []batching.Job, jobResult chan<- batching.JobResult) []batching.JobResult {
	// < processing of the batch happen here >
	// just add some random sleep for processing time
	randomSleepTime := time.Millisecond * time.Duration(rand.Intn(1000-100)+100)
	time.Sleep(randomSleepTime)

	// Notify that jobs in batch successfully processed
	var jobResults []batching.JobResult
	for _, job := range jobs {
		result := batching.JobResult{ID: job.ID, Success: true}
		jobResults = append(jobResults, result)

		// write result in to channel
		jobResult <- result
	}
	return jobResults
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	batch := batching.NewBatching[batching.Job](ctx, 10, time.Millisecond*500)

	p := processor{}

	go func() {
		for {
			// Consume batches
			batchToProcess := <-batch.Batch
			// Processed batches
			results := p.BatchProcessor(batchToProcess, batch.JobResult)
			fmt.Printf("batch size of %d processed\n", len(results))
		}
	}()

	// If we need to communicate JobResult one by one, use <-batch.JobResult channel
	go func() {
		for {
			result := <-batch.JobResult
			fmt.Printf("result Job ID: %d proccesed %v\n", result.ID, result.Success)
		}
	}()

	// Populate test data
	for i := 1; i <= 112; i++ {
		randomSleepTime := time.Millisecond * time.Duration(rand.Intn(100-10)+10)
		time.Sleep(randomSleepTime)
		// set Jobs
		batch.Job <- batching.Job{ID: i, Payload: nil}
	}

	// shutdown method
	cancel()
}

```