package hw05parallelexecution

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

var ErrInvalidData = errors.New("invalid data: n and m must be > 0")

type Task func() error

type counter struct {
	mu  sync.RWMutex
	val int
}

func (c *counter) increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.val++
}

func (c *counter) value() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.val
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
					errorCount.increment()
				}
				wg.Done()
			}
		}()
	}
	for _, task := range tasks {
		if errorCount.value() >= m {
			break
		}
		wg.Add(1)
		ch <- task
	}
	close(ch)
	wg.Wait()
	if errorCount.value() >= m {
		return ErrErrorsLimitExceeded
	}
	return nil
}
