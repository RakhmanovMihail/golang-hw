package hw06pipelineexecution

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	sleepPerStage = time.Millisecond * 100
	fault         = sleepPerStage / 2
)

func TestPipeline(t *testing.T) {
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("simple case", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t,
			int64(elapsed),
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Len(t, result, 0)
		require.Less(t, int64(elapsed), int64(abortDur)+int64(fault))
	})

	t.Run("empty input", func(t *testing.T) {
		in := make(Bi)
		close(in) // сразу закрыли вход

		var result []interface{}
		for v := range ExecutePipeline(in, nil, stages...) {
			result = append(result, v)
		}

		require.Empty(t, result, "pipeline should return empty result on closed input")
	})

	t.Run("no stages", func(t *testing.T) {
		in := make(Bi)
		data := []int{42, 100}
		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		var result []interface{}
		for v := range ExecutePipeline(in, nil) {
			result = append(result, v)
		}

		require.Equal(t, []interface{}{42, 100}, result)
	})

	t.Run("single stage", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3}
		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		var result []int
		for v := range ExecutePipeline(in, nil, stages[1]) {
			result = append(result, v.(int))
		}

		require.Equal(t, []int{2, 4, 6}, result)
	})

	t.Run("immediate done", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		close(done)

		go func() {
			in <- 1
			in <- 2
			close(in)
		}()

		var result []interface{}
		for v := range ExecutePipeline(in, done, stages...) {
			result = append(result, v)
		}

		require.Empty(t, result, "pipeline should stop immediately if done is closed")
	})
}
