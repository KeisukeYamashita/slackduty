package slackduty

import (
	"context"

	"github.com/robfig/cron"
	"go.uber.org/zap"
)

// Job represents a syncronize execution between
// Slack and Pagerduty resources.
type Job interface {
	Run(context.Context) error
}

type job struct {
	fn     func() error
	logger *zap.Logger
}

type cronJob struct {
	cron     *cron.Cron
	fn       func() error
	schedule string
	logger   *zap.Logger
}

type jobOptions struct {
	cron            *cron.Cron
	schedule        string
	externalTrigger bool
}

var defaultJobOptions = jobOptions{externalTrigger: true}

// JobOption configures the job
type JobOption func(*jobOptions)

// WithSchedule is intened to run the job by cron.
// Pass the same cron as you call the start function.
// Schedule format should follow github.com/robfig/cron
func WithSchedule(goCron *cron.Cron, schedule string) JobOption {
	return func(o *jobOptions) {
		o.cron = goCron
		o.schedule = schedule
		o.externalTrigger = false
	}
}

// NewJob creates a new (cron)job.
// Run() method will run the job returing the error.
func NewJob(fn func() error, opts ...JobOption) Job {
	var o jobOptions
	o = defaultJobOptions
	for _, opt := range opts {
		opt(&o)
	}

	if o.externalTrigger {
		return &job{
			fn: fn,
		}
	}

	return &cronJob{
		fn:       fn,
		schedule: o.schedule,
		cron:     o.cron,
	}
}

// Run a cronjob based on cron schedule.
// To cancle a job, pass the a context.
func (cj *cronJob) Run(ctx context.Context) error {
	errChan := make(chan error, 1)
	go func() {
		cj.cron.AddFunc(cj.schedule, func() {
			errChan <- cj.fn()
		})
	}()

	for {

	}

	return nil
}

// Run a single job than finished after one execution.
func (j *job) Run(ctx context.Context) error {
	return j.fn()
}
