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

// ErrJobStatusTransitionNotAllowed is an error that indicates an invalid job
// status transition
type ErrJobStatusTransitionNotAllowed struct {
	JobID int
}

func (e ErrJobStatusTransitionNotAllowed) Error() string {
	return fmt.Sprintf("unable to transition job status for job %d", e.JobID)
}
