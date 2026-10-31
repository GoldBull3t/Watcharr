// We will use jobs only for storing/retrieving active job statuses for the client.
// Running the job will be done wherever needed, but isn't handled here.
// When starting a job elsewhere, we should first add a job here as active to get an `id`,
// this id should be used to update the active job so the client can request job status updates.

package job

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/sbondCo/Watcharr/util"
)

type JobStatus string

var (
	JOB_CREATED   JobStatus = "CREATED"
	JOB_RUNNING   JobStatus = "RUNNING"
	JOB_DONE      JobStatus = "DONE"
	JOB_CANCELLED JobStatus = "CANCELLED"
)

type Job struct {
	// We can give the job a name simply for showing on the client.
	Name string `json:"name"`
	// The current status of the job.
	Status JobStatus `json:"status"`
	// The current task we are performing inside the job.
	// Just so we can portray progress on the client by displaying the current task.
	CurrentTask string `json:"currentTask,omitempty"`
	// Errors that occurred in the task
	Errors []string `json:"errors"`
	// Stored for access control.
	UserId uint `json:"-"`
}

var (
	jobsMu     sync.RWMutex
	activeJobs = make(map[string]*Job)
	stopOnce   sync.Once
	stopCh     = make(chan struct{})
)

// ShutdownJobs signals any waiting cleanup timers to terminate immediately.
func ShutdownJobs() {
	stopOnce.Do(func() {
		close(stopCh)
	})
}

// Add a job to our activeJobs map.
// Returns id of job on success, or error if failed to add.
// Only return safe errors for display to users, log serious errors.
func AddJob(name string, userId uint) (string, error) {
	idk, err := util.GenerateString(8)
	if err != nil {
		slog.Error("addJob: Failed to generate a job id!", "error", err)
		return "", errors.New("failed to generate a job id, please try again")
	}
	jobsMu.Lock()
	defer jobsMu.Unlock()
	_, ok := activeJobs[idk]
	if ok {
		// Lets just hope this doesn't happen, may the odds be with us.
		return "", errors.New("job already exists with id generated, please try again")
	}
	activeJobs[idk] = &Job{
		Name:   name,
		Status: JOB_CREATED,
		UserId: userId,
	}
	return idk, nil
}

// Add a job, but only if one with the same `name` isn't already running.
// Only return safe errors for display to users.
func AddUniqueJob(name string, userId uint) (string, error) {
	jobsMu.RLock()
	found := false
	for _, v := range activeJobs {
		if v.UserId == userId && v.Name == name && (v.Status == JOB_CREATED || v.Status == JOB_RUNNING) {
			found = true
			break
		}
	}
	jobsMu.RUnlock()
	if found {
		return "", errors.New("a job of this type is already running, please wait for the existing job to finish")
	}
	return AddJob(name, userId)
}

func RmJob(id string, userId uint) {
	slog.Debug("rmJob: Removing a job.", "id", id)
	jobsMu.Lock()
	defer jobsMu.Unlock()
	v, ok := activeJobs[id]
	if ok && v.UserId == userId {
		delete(activeJobs, id)
		slog.Debug("rmJob: Removed a job.", "id", id)
		return
	}
	slog.Debug("rmJob: Job to remove does not exist (or not owned by this user).", "id", id, "user_id", userId)
}

// Get a job.
// Returns job if found, otherwise errors if job does not exist.
func GetJob(id string, userId uint) (*Job, error) {
	jobsMu.RLock()
	defer jobsMu.RUnlock()
	j, ok := activeJobs[id]
	if ok {
		// Ensure user requesting a job, owns the job.
		if j.UserId != userId {
			slog.Warn("getJob: A user tried to access a job they do not own.", "user_id", userId, "job_id", id)
			return &Job{}, errors.New("job does not exist")
		}
		// Return a shallow copy of the job so caller reads are safe
		jobCopy := *j
		return &jobCopy, nil
	}
	return &Job{}, errors.New("job does not exist")
}

// Update a jobs status.
func UpdateJobStatus(id string, userId uint, status JobStatus) error {
	jobsMu.Lock()
	j, ok := activeJobs[id]
	if !ok || j.UserId != userId {
		jobsMu.Unlock()
		slog.Error("updateJobStatus: Failed!", "status", status, "error", "job does not exist")
		return errors.New("job does not exist")
	}
	j.Status = status
	jobsMu.Unlock()

	// If job is set to done, remove it after 30 minutes.
	if status == JOB_DONE || status == JOB_CANCELLED {
		slog.Debug("updateJobStatus: Job set to done or cancelled. Will be removed after 30m.", "id", id, "status", status)
		go func() {
			t := time.NewTimer(30 * time.Minute)
			defer t.Stop()
			select {
			case <-t.C:
			case <-stopCh:
				return
			}
			slog.Debug("updateJobStatus: Job done. waited 30m.. removing job now.", "id", id)
			RmJob(id, userId)
		}()
	}
	return nil
}

// Update a jobs current task.
func UpdateJobCurrentTask(id string, userId uint, ct string) error {
	jobsMu.Lock()
	defer jobsMu.Unlock()
	j, ok := activeJobs[id]
	if !ok || j.UserId != userId {
		slog.Error("updateJobCurrentTask: Failed!", "ct", ct, "error", "job does not exist")
		return errors.New("job does not exist")
	}
	j.CurrentTask = ct
	return nil
}

// Add an error to a job.
func AddJobError(id string, userId uint, e string) error {
	jobsMu.Lock()
	defer jobsMu.Unlock()
	j, ok := activeJobs[id]
	if !ok || j.UserId != userId {
		slog.Error("addJobError: Failed!", "e", e, "error", "job does not exist")
		return errors.New("job does not exist")
	}
	j.Errors = append(j.Errors, e)
	return nil
}
