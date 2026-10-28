package task

import (
	"errors"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database"
	"github.com/sbondCo/Watcharr/feature/arr"
	"github.com/sbondCo/Watcharr/image"
	"github.com/sbondCo/Watcharr/token"
	"gorm.io/gorm"
)

type TaskRescheduleRequest struct {
	// Number of seconds inbetween each run of this task.
	Seconds int `json:"seconds" binding:"required"`
}

type AllTasksResponse struct {
	// The tasks name.
	Name string `json:"name"`
	// When this task will next run.
	NextRun time.Time `json:"nextRun"`
	// Current schedule for this task (seconds).
	Seconds int `json:"seconds"`
}

type TaskFunc struct {
	// Task function.
	f func()
	// Default duration (schedule) for task.
	dd time.Duration
}

var taskScheduler gocron.Scheduler

// All task functions are stored here so when updating (rescheduling)
// a job, we can give it this function again.
// Doesn't seem to be a way to only update the schedule of a job,
// the .Update func wants the whole definition again.
//
// All funcs simply call a cleaning/routine method where the rest of the
// related code lives so it's kept tidy.
var taskFuncs map[string]TaskFunc

type CalendarNotifier interface {
	CheckAndNotifyAllUsers() error
}

type BackupProvider interface {
	RunAutomatedBackup() error
}

// Setup recurring tasks (eg cleanup every x mins)
func SetupTasks(cfg *config.ServerConfig, db *gorm.DB, calendarNotifier CalendarNotifier, backupProvider ...BackupProvider) {
	ts, err := gocron.NewScheduler()
	if err != nil {
		slog.Error("SetupTasks: Failed to create new scheduler!", "error", err)
		return
	}
	taskScheduler = ts

	// Define all task funcs.
	taskFuncs = map[string]TaskFunc{
		"Cleanup Tokens": {
			f: func() {
				token.CleanupTokens(db)
			},
			dd: 60 * time.Second,
		},
		"Refresh Arr Queues": {
			f: func() {
				arr.RefreshArrQueues(cfg)
			},
			dd: 60 * time.Second,
		},
		"Cleanup Images": {
			f: func() {
				image.CleanupImages(db)
			},
			dd: 24 * time.Hour,
		},
		"Optimize Database": {
			f: func() {
				database.TaskOptimize(db)
			},
			dd: 24 * time.Hour,
		},
		"Memory Scavenger": {
			f: func() {
				if db != nil && db.Dialector.Name() == "sqlite" {
					db.Exec("PRAGMA shrink_memory")
				}
				debug.FreeOSMemory()
				slog.Debug("Memory Scavenger: Returned unused memory pages to OS")
			},
			dd: 3 * time.Minute,
		},
	}

	if calendarNotifier != nil {
		taskFuncs["Calendar Release Notifications"] = TaskFunc{
			f: func() {
				slog.Debug("Running Calendar Release Notifications recurring task")
				_ = calendarNotifier.CheckAndNotifyAllUsers()
			},
			dd: 6 * time.Hour,
		}
	}

	if len(backupProvider) > 0 && backupProvider[0] != nil {
		bp := backupProvider[0]
		taskFuncs["Automated Database Backup"] = TaskFunc{
			f: func() {
				slog.Debug("Running Automated Database Backup recurring task")
				if err := bp.RunAutomatedBackup(); err != nil {
					slog.Error("Automated Database Backup failed", "error", err)
				}
			},
			dd: 24 * time.Hour,
		}
	}

	// Add all jobs to scheduler.
	for k, v := range taskFuncs {
		err = AddTaskToScheduler(cfg, k, v.dd)
		if err != nil {
			slog.Error("SetupTasks: Failed to add new job", "job", k, "err", err)
		}
	}

	taskScheduler.Start()
	slog.Info("SetupTasks: Jobs created and scheduler started.")
}

// Gets schedule from config, or `defaultDur` if not manually configured.
func GetTaskSeconds(cfg *config.ServerConfig, name string, defaultDur time.Duration) time.Duration {
	s := defaultDur
	if sec := cfg.GetTaskSchedule(name); sec != 0 {
		s = time.Duration(sec) * time.Second
	}
	return s
}

// ShutdownTasks gracefully stops the task scheduler.
func ShutdownTasks() {
	if taskScheduler != nil {
		if err := taskScheduler.Shutdown(); err != nil {
			slog.Warn("ShutdownTasks: Failed to shutdown scheduler cleanly", "error", err)
		} else {
			slog.Info("ShutdownTasks: Scheduler stopped cleanly.")
		}
	}
}

// Add new job to scheduler.
func AddTaskToScheduler(cfg *config.ServerConfig, name string, defaultDur time.Duration) error {
	s := GetTaskSeconds(cfg, name, defaultDur)
	_, err := taskScheduler.NewJob(
		gocron.DurationJob(s),
		gocron.NewTask(taskFuncs[name].f),
		gocron.WithName(name),
	)
	slog.Debug("addTaskToScheduler: Job added.", "job_name", name, "duration_used", s, "duration_default", defaultDur)
	return err
}

// Get all tasks in a consumable format.
func GetAllTasks(cfg *config.ServerConfig) []AllTasksResponse {
	jobs := []AllTasksResponse{}
	for _, j := range taskScheduler.Jobs() {
		j2a := AllTasksResponse{
			Name: j.Name(),
		}
		nextRun, err := j.NextRun()
		if err != nil {
			slog.Error("getAllTasks: Failed to get next run time for a job.", "job_name", j2a.Name)
		} else {
			j2a.NextRun = nextRun
		}
		j2a.Seconds = int(GetTaskSeconds(cfg, j2a.Name, taskFuncs[j2a.Name].dd).Seconds())
		jobs = append(jobs, j2a)
	}
	return jobs
}

// Get task (job) from scheduler by name.
func GetTask(name string) *gocron.Job {
	for _, j := range taskScheduler.Jobs() {
		if j.Name() == name {
			job := j
			return &job
		}
	}
	return nil
}

// Reschedule a task by name.
func RescheduleTask(cfg *config.ServerConfig, name string, req TaskRescheduleRequest) error {
	if req.Seconds == 0 {
		return errors.New("request has no seconds")
	}
	j := GetTask(name)
	if j == nil {
		return errors.New("no task found")
	}
	// Update config
	cfg.SetTaskSchedule(name, req.Seconds)
	if err := cfg.Write(); err != nil {
		slog.Error("rescheduleTask: Failed to write updated config to file!", "error", err)
		return errors.New("failed to write config")
	}
	// Update job in scheduler
	_, err := taskScheduler.Update(
		(*j).ID(),
		gocron.DurationJob(
			time.Duration(req.Seconds)*time.Second,
		),
		gocron.NewTask(
			taskFuncs[name].f,
		),
		gocron.WithName(name),
	)
	if err != nil {
		slog.Error("rescheduleTask: Failed to update job!", "error", err)
		return errors.New("failed to update job")
	}
	return nil
}
