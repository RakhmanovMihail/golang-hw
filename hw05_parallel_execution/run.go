package hw05parallelexecution

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")
var ErrInvalidData = errors.New("invalid data: n and m must be > 0")

type Task func() error

type counter struct {
	sync.Mutex
	value int
}

func (c *counter) Increment() {
	c.Lock()
	defer c.Unlock()
	c.value++
}
func (c *counter) Value() int {
	c.Lock()
	defer c.Unlock()
	return c.value
}

func Run(tasks []Task, n, m int) error {
	if n <= 0 || m <= 0 {
		return ErrInvalidData
	}
	if n >= len(tasks) {
		n = len(tasks)
	}
	if len(tasks) == 0 {
		return nil
	}

	var errorCount counter
	var wg sync.WaitGroup
	ch := make(chan Task)

	for i := 0; i < n; i++ {
		go func() {
			for task := range ch {
				err := task()
				if err != nil {
					errorCount.Increment()
				}
				wg.Done()
			}
		}()
	}

	for _, task := range tasks {
		if errorCount.Value() >= m {
			break
		}
		wg.Add(1)
		ch <- task
	}
	close(ch)

	wg.Wait()
	if errorCount.Value() >= m {
		return ErrErrorsLimitExceeded
	}
	return nil
}
