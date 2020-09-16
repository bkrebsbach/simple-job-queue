package domain

const (
	JobTypeTimeCritical    = "TIME_CRITICAL"
	JobTypeNotTimeCritical = "NOT_TIME_CRITICAL"

	JobStatusQueued     = "QUEUED"
	JobStatusInProgress = "IN_PROGRESS"
	JobStatusConcluded  = "CONCLUDED"
	JobStatusCancelled  = "CANCELLED"
)

var (
	// JobStatues defines valide job status values
	JobStatuses = map[string]bool{
		JobStatusQueued:     true,
		JobStatusInProgress: true,
		JobStatusConcluded:  true,
	}

	// JobTypes defines valid job type values
	JobTypes = map[string]bool{
		JobTypeTimeCritical:    true,
		JobTypeNotTimeCritical: true,
	}
)

// Job defines the basic job structure
type Job struct {
	ID         int
	Type       string
	Status     string
	ConsumerID string
}
