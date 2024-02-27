package batching

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestNumberOfBatches(t *testing.T) {
	batch := NewBatching[Job](context.TODO(), 100, time.Millisecond*100)
	go func() {
		for i := 1; i <= 500; i++ {
			batch.Job <- Job{ID: i}
		}
	}()
	var count int
	go func() {
		for {
			<-batch.Batch
			count++
		}
	}()
	time.Sleep(time.Second * 1)
	if count != 5 {
		t.Fatalf("invalid batch count: %d", count)
	}
}

func TestBatchSize(t *testing.T) {
	batch := NewBatching[Job](context.TODO(), 10, time.Second*1)
	go func() {
		for i := 1; i <= 50; i++ {
			batch.Job <- Job{ID: i}
		}
	}()
	batchSize := reflect.ValueOf(<-batch.Batch).Len()
	if batchSize != 10 {
		t.Fatalf("invalid batch size: %d", batchSize)
	}
}

func TestBatchMaxWait(t *testing.T) {
	batch := NewBatching[Job](context.TODO(), 100, time.Second*1)
	go func() {
		for i := 1; i <= 10; i++ {
			batch.Job <- Job{ID: i}
		}
		time.Sleep(time.Second * 2)
		for i := 1; i <= 20; i++ {
			batch.Job <- Job{ID: i}
		}
	}()

	// fist batch trigger in one second with size 10
	batchSize := reflect.ValueOf(<-batch.Batch).Len()
	if batchSize != 10 {
		t.Fatalf("invalid batch size: %d", batchSize)
	}
	// next batch should be 20
	batchSize = reflect.ValueOf(<-batch.Batch).Len()
	if batchSize != 20 {
		t.Fatalf("invalid batch size: %d", batchSize)
	}
}

func TestBatchWriteToClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	batch := NewBatching[Job](ctx, 100, time.Millisecond*1)

	cancel()

	// could not write into closed
	go func() {
		for i := 1; i <= 10; i++ {
			batch.Job <- Job{ID: i}
		}
	}()

	time.Sleep(time.Millisecond * 10)

	batchSize := reflect.ValueOf(<-batch.Batch).Len()
	if batchSize != 0 {
		t.Fatalf("invalid batch size: %d", batchSize)
	}
}

func TestBatchReadFromClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	batch := NewBatching[Job](ctx, 10, time.Millisecond*1)

	// could not write into closed
	go func() {
		for i := 1; i <= 10; i++ {
			batch.Job <- Job{ID: i}
		}
	}()

	time.Sleep(time.Millisecond * 20)
	cancel()

	// consume the rest after close
	batchSize := reflect.ValueOf(<-batch.Batch).Len()
	if batchSize != 10 {
		t.Fatalf("invalid batch size: %d", batchSize)
	}
}
