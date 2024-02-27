package batching

import (
	"context"
	"time"
)

type Job struct {
	ID      int
	Payload interface{}
}

type JobResult struct {
	ID      int
	Success bool
}

type options struct {
	batchSize int           // max batch size
	maxWait   time.Duration // maximum waiting time for a max batch.
	ctx       context.Context
}

type Batch[T any] struct {
	opts      *options
	Job       chan<- T       // Job represents the batch input channel.
	Batch     <-chan []T     // Batch represents the batch output channel.
	JobResult chan JobResult // Job Result
}

func NewBatching[T any](ctx context.Context, batchSize int, batchInterval time.Duration) *Batch[T] {
	opts := &options{
		batchSize: batchSize,
		maxWait:   batchInterval,
		ctx:       ctx,
	}

	job := make(chan T)
	batchOutput := make(chan []T)
	jobResult := make(chan JobResult)

	b := &Batch[T]{
		opts:      opts,
		Job:       job,
		Batch:     batchOutput,
		JobResult: jobResult,
	}

	go b.batching(job, batchOutput)

	return b
}

// this function is responsible to create batches from jobs
func (b *Batch[T]) batching(job chan T, batchOutput chan []T) {
	buffer := make([]T, 0, b.opts.batchSize)
	ticker := time.NewTicker(b.opts.maxWait)

Loop:
	for {
		if b.opts.ctx.Err() != nil {
			break
		}

		select {
		case event := <-job:
			buffer = append(buffer, event)

			if len(buffer) == b.opts.batchSize {
				batchOutput <- buffer
				buffer = make([]T, 0, b.opts.batchSize)
				ticker.Reset(b.opts.maxWait)
			}

		case <-ticker.C:
			if len(buffer) > 0 {
				batchOutput <- buffer
				buffer = make([]T, 0, b.opts.batchSize)
				ticker.Reset(b.opts.maxWait)
			}

		case <-b.opts.ctx.Done():
			break Loop
		}
	}

	if len(buffer) > 0 {
		batchOutput <- buffer
	}

	close(batchOutput)
}
