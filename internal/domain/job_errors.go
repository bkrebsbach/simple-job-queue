package domain

import (
	"errors"
	"fmt"
)

var (
	ErrQueueEmpty = errors.New("no jobs in queue")
)

type ErrJobNotFound struct {
	JobID int
}

func (e ErrJobNotFound) Error() string {
	return fmt.Sprintf("unable to find job %d", e.JobID)
}
