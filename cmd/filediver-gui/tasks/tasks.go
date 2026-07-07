package tasks

import (
	"context"
	"sync"
)

type TaskFunc func(ctx context.Context, params map[string]any, onProgress func(prog float64), onStatus func(string)) (result map[string]any, err error)

type TaskState struct {
	// TaskFunc used to run the task.
	Fn TaskFunc
	// Status string.
	Status string
	// Progress in range [0,1].
	Prog float64
	// True when the task has finished (regardless of success).
	Done bool
	// Set to true for one call of [TaskExecution.Poll] just
	// after the task has finished, regardless of success
	// (use [TaskExecution.SoftPoll] to not affect this field
	// when polling).
	JustFinished bool
	// Task result and error (only set after task has finished).
	// Err is set to [context.Canceled] if the task was canceled
	// prematurely.
	Res map[string]any
	Err error
}

type TaskExecution struct {
	lock   sync.Mutex
	state  TaskState
	cancel func(error)
}

// Returns a new [TaskExecution], which has just finished with the given error.
func NewErroredTaskExecution(err error) *TaskExecution {
	return &TaskExecution{state: TaskState{
		Err:          err,
		Done:         true,
		JustFinished: true,
	}}
}

// Returns a snapshot of the current execution status.
//
// If ex is nil, returns a [TaskExecution] with Done set to true and all other fields zeroed.
func (ex *TaskExecution) Poll() TaskState {
	if ex == nil {
		return TaskState{Done: true}
	}
	var st TaskState
	ex.lock.Lock()
	st = ex.state
	ex.state.JustFinished = false
	ex.lock.Unlock()
	return st
}

// Like [TaskExecution.Poll], but doesn't reset JustFinished.
func (ex *TaskExecution) SoftPoll() TaskState {
	if ex == nil {
		return TaskState{Done: true}
	}
	var st TaskState
	ex.lock.Lock()
	st = ex.state
	ex.lock.Unlock()
	return st
}

// Like [TaskExecution.Cancel], but with an explicit cause.
func (ex *TaskExecution) CancelCause(cause error) {
	if ex != nil && ex.cancel != nil {
		ex.cancel(cause)
	}
}

// Cancels the underlying cancellable context.
//
// Does not need to be called manually after the task has finished.
//
// Does nothing if ex is nil. May be called multiple times and from any goroutine.
func (ex *TaskExecution) Cancel() {
	ex.CancelCause(nil)
}

func (t TaskFunc) executeCommon(outCh chan<- TaskState, ctx context.Context, params map[string]any) (_ *TaskExecution, run func()) {
	cctx, cancel := context.WithCancelCause(ctx)
	ex := &TaskExecution{
		state:  TaskState{Fn: t},
		cancel: cancel,
	}
	return ex, func() {
		res, err := t(cctx, params, func(prog float64) {
			ex.lock.Lock()
			ex.state.Prog = min(max(prog, 0), 1)
			st := ex.state
			ex.lock.Unlock()
			if outCh != nil {
				outCh <- st
			}
		}, func(s string) {
			ex.lock.Lock()
			ex.state.Status = s
			st := ex.state
			ex.lock.Unlock()
			if outCh != nil {
				outCh <- st
			}
		})
		ex.lock.Lock()
		ex.state.Err = err
		if err == nil {
			ex.state.Res = res
		}
		ex.state.Prog = 1
		ex.state.Done = true
		ex.state.JustFinished = true
		st := ex.state
		ex.lock.Unlock()
		if outCh != nil {
			outCh <- st
		}
		cancel(nil)
	}
}

// Runs the task synchronously.
func (t TaskFunc) Run(ctx context.Context, params map[string]any) TaskState {
	ex, run := t.executeCommon(nil, ctx, params)
	run()
	return ex.Poll()
}

// Runs the task asynchronously (with polling).
func (t TaskFunc) Go(ctx context.Context, params map[string]any) *TaskExecution {
	ex, run := t.executeCommon(nil, ctx, params)
	go run()
	return ex
}

// Runs the task asynchronously with the given channel. A snapshot of the state
// is sent on every update. The channel is closed after the task has completed.
func (t TaskFunc) GoCh(outCh chan<- TaskState, ctx context.Context, params map[string]any) *TaskExecution {
	ex, run := t.executeCommon(outCh, ctx, params)
	go run()
	return ex
}
