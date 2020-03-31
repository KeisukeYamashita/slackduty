package slackduty

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/robfig/cron"
)

func TestWithSchedule(t *testing.T) {
	type input struct {
		cron     *cron.Cron
		schedule string
	}

	goCron := cron.New()

	tcs := map[string]struct {
		input *input
	}{
		"pass": {&input{
			cron:     goCron,
			schedule: "* * * * *",
		}},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			opts := []JobOption{WithSchedule(tc.input.cron, tc.input.schedule)}
			var o jobOptions
			for _, opt := range opts {
				opt(&o)
			}

			if !reflect.DeepEqual(o.cron, tc.input.cron) {
				t.Fatal("job option cron doesn't match")
			}

			if o.externalTrigger {
				t.Fatal("job option externalTrigger should be false")
			}

			if o.schedule != tc.input.schedule {
				t.Fatalf("job option schedule doesn't match got: %s want: %s", o.schedule, tc.input.schedule)
			}
		})
	}

}

func TestNewJob(t *testing.T) {
	goCron := cron.New()

	tcs := map[string]struct {
		opts            []JobOption
		externalTrigger bool
	}{
		"pass": {
			externalTrigger: true,
			opts:            []JobOption{},
		},
		"pass with schedule": {
			externalTrigger: false,
			opts:            []JobOption{WithSchedule(goCron, "* * * * *")},
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			got := NewJob(func() error { return nil }, tc.opts...)

			switch v := got.(type) {
			case *job:
			case *cronJob:
				if !reflect.DeepEqual(v.cron, goCron) {
					diff := cmp.Diff(v.cron, goCron)
					t.Fatalf("job shouldn't be cronjob diff: %v", diff)
				}

				if v.schedule != "* * * * *" {
					t.Fatalf("schedule not matched got: %s", v.schedule)
				}
			}
		})
	}
}

func TestRun_CronJob(t *testing.T) {
	var (
		testSchedule = "* * * * *"
		goCron       = cron.New()
	)

	tcs := map[string]struct {
		fn      func() error
		ctx     context.Context
		success bool
	}{
		"pass": {func() error { return nil }, context.Background(), true},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			opt := WithSchedule(goCron, testSchedule)
			cronjob := NewJob(tc.fn, opt)
			ctx, cancel := context.WithCancel(tc.ctx)
			defer cancel()
			runctx, cancel2 := context.WithCancel(ctx)
			errChan := make(chan error, 1)

			go func() {
				errChan <- cronjob.Run(runctx)
			}()

			cancel2()

			select {
			case <-ctx.Done():
				t.Fatalf("canceled: %v", ctx.Err())
			case <-errChan:
			}
		})
	}
}

func TestRun_Job(t *testing.T) {
	tcs := map[string]struct {
		fn      func() error
		ctx     context.Context
		success bool
	}{
		"pass": {func() error { return nil }, context.Background(), true},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			job := NewJob(tc.fn)
			err := job.Run(tc.ctx)
			if err != nil && tc.success {
				t.Fatalf("test error: %v", err)
			}
		})
	}
}
