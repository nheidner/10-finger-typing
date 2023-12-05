package utils

import "time"

func Throttle(interval time.Duration) (execute func(execFunc func()), cleanup func()) {
	executeCh := make(chan func(), 1)
	ticker := time.NewTicker(interval)

	go func() {
		var nextFunc func()
		for {
			select {
			case f := <-executeCh:
				nextFunc = f
			case <-ticker.C:
				if nextFunc != nil {
					nextFunc()
					nextFunc = nil
				}
			}
		}
	}()

	cleanup = func() {
		ticker.Stop()
		close(executeCh)
	}

	execute = func(execFunc func()) {
		executeCh <- execFunc
	}

	return execute, cleanup
}
