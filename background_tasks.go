package batt

import (
	"log"
	"time"

	"github.com/alitto/pond/v2"
)

var (
	OnProcessTaskError = func(err error) { log.Println(err) }
	pool               pond.Pool
)

type TaskFunc func() error

// Creates a task worker pool with the specified maximum concurrency.
func InitTaskWorkerPool(maxWorkers int) {
	pool = pond.NewPool(maxWorkers)
}

// Stops the pool and waits for all tasks to complete.
func StopAndWaitWorkerPool() {
	pool.StopAndWait()
}

// ProcessTask keeps trying the function until the second argument
// returns false, or no error is returned.
func ProcessTask(fn TaskFunc, retries int, delay time.Duration) {
	var err error

	pool.Submit(func() {
		for range retries {
			err = fn()
			if err == nil {
				return
			}

			time.Sleep(delay)
		}
	})

	OnProcessTaskError(err)
}
