package sched

import (
	"fmt"
	"reflect"
	"time"

	"github.com/kataras/golog"
)

// DefaultLocation is the default location that the `FromDate` helper
// uses, it's the `time.Local` but you can change it to whatever fills your needs, i.e `time.UTC`
// or `time.LoadLocation("Europe/Athens")` for example.
var DefaultLocation = time.Local

// FromDate takes a date and calculates the `time.Duration` from `time.Now()`.
// The date should be after the `time.Now()`.
func FromDate(year int, month time.Month, day, hour, min, sec, nsec int) time.Duration {
	date := time.Date(year, month, day, hour, min, sec, nsec, DefaultLocation)
	now := time.Now()
	if date.Before(now) {
		golog.Errorf("scheduler: can't schedule something that happened in the past(%s)", date.String())
		return 0
	}
	return date.Sub(now)
}

// CancelFunc is the type of the Schedule's cancel function.
// It cancels the scheduled job by calling the timer's `Stop`.
// If cancel functions returns false, then the timer
// has already expired and the job has been started in its own goroutine.
type CancelFunc func() bool

var (
	emptyIn     = []reflect.Value{}
	emptyCancel = func() bool { return false }
)

// Schedule schedules a job function with any optional job's function input arguments.
// The time.Duration is the given time point that this job will be executed, you can
// use helpers like the `FromDate` to calculate the time.Duration or just pass a strict
// duration like `24 *time.Hour`.
//
// It returns two values, the first one is the cancel function, if called then the job is canceled
// and returns true, otherwise (if it's already stopped) it returns false.
// The last is an error value which is not nil if any error occurred.
func Schedule(when time.Duration, job interface{}, jobInput ...interface{}) (cancel CancelFunc, err error) {
	// minimum nanoseconds.
	if when <= 60 {
		err := fmt.Errorf("scheduler: very low 'when' duration(%s) passed", when.String())
		golog.Error(err)
		return emptyCancel, err
	}

	in := emptyIn

	if n := len(jobInput); n > 0 {
		in = make([]reflect.Value, n, n)
		for i, actionIn := range jobInput {
			v := reflect.ValueOf(actionIn)
			if !v.IsValid() {
				err := fmt.Errorf("scheduler: job's input argument: %v is invalid", actionIn)
				golog.Error(err)
				return emptyCancel, err
			}
			in[i] = v
		}
	}

	fn := reflect.ValueOf(job)
	if fn.Kind() != reflect.Func {
		err := fmt.Errorf("scheduler: 'job' is not a function")
		golog.Error(err)
		return emptyCancel, err
	}

	t := time.AfterFunc(when, func() {
		fn.Call(in)
	})

	return t.Stop, nil
}
