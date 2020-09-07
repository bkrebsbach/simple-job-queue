package domain

import (
	"errors"
	"fmt"
)

var (
	// ErrQueueEmpty is an error indicating there are no jobs in the queue.
	ErrQueueEmpty = errors.New("no jobs in queue")
)

// ErrJobNotFound indicates a given job ID was not found in the queue.
type ErrJobNotFound struct {
	JobID int
}

func (e ErrJobNotFound) Error() string {
	return fmt.Sprintf("unable to find job %d", e.JobID)
}
