// Package task manage task, such as file upload, file copy between storages, offline download, etc.
package task

import (
	"context"
	"runtime"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	PENDING   = "pending"
	RUNNING   = "running"
	SUCCEEDED = "succeeded"
	CANCELING = "canceling"
	CANCELED  = "canceled"
	ERRORED   = "errored"
)

type Func[K comparable] func(task *Task[K]) error
type Callback[K comparable] func(task *Task[K])

type Task[K comparable] struct {
	ID       K
	Name     string
	state    string // pending, running, finished, canceling, canceled, errored
	status   string
	progress float64

	Error error

	Func     Func[K]
	callback Callback[K]

	Ctx    context.Context
	cancel context.CancelFunc

	mu    sync.RWMutex
	runMu sync.Mutex
}

func (t *Task[K]) SetStatus(status string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.status = status
}

func (t *Task[K]) SetProgress(percentage float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.progress = percentage
}

func (t *Task[K]) GetProgress() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.progress
}

func (t *Task[K]) GetState() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state
}

func (t *Task[K]) GetStatus() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

func (t *Task[K]) GetError() error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Error
}

func (t *Task[K]) GetErrMsg() string {
	err := t.GetError()
	if err == nil {
		return ""
	}
	return err.Error()
}

func getCurrentGoroutineStack() string {
	buf := make([]byte, 1<<16)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

func (t *Task[K]) run() {
	t.runMu.Lock()
	defer t.runMu.Unlock()
	t.mu.Lock()
	t.state = RUNNING
	t.mu.Unlock()
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("error [%s] while run task [%s],stack trace:\n%s", err, t.Name, getCurrentGoroutineStack())
			t.mu.Lock()
			t.Error = errors.Errorf("panic: %+v", err)
			t.state = ERRORED
			t.mu.Unlock()
		}
	}()
	err := t.Func(t)
	t.mu.Lock()
	t.Error = err
	t.mu.Unlock()
	if err != nil {
		log.Errorf("error [%+v] while run task [%s]", err, t.Name)
	}

	t.mu.Lock()
	canceled := errors.Is(t.Ctx.Err(), context.Canceled) || t.state == CANCELING
	succeeded := false
	if canceled {
		t.state = CANCELED
	} else if err != nil {
		t.state = ERRORED
	} else {
		t.state = SUCCEEDED
		succeeded = true
	}
	callback := t.callback
	t.mu.Unlock()

	if succeeded {
		t.SetProgress(100)
		if callback != nil {
			callback(t)
		}
	}
}

func (t *Task[K]) retry() {
	t.run()
}

func (t *Task[K]) Done() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state == SUCCEEDED || t.state == CANCELED || t.state == ERRORED
}

func (t *Task[K]) Cancel() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.state == SUCCEEDED || t.state == CANCELED {
		return
	}
	if t.cancel != nil {
		t.cancel()
	}
	// maybe can't cancel
	t.state = CANCELING
}

func (t *Task[K]) markCanceled() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.state != SUCCEEDED && t.state != ERRORED {
		t.state = CANCELED
	}
}

func WithCancelCtx[K comparable](task *Task[K]) *Task[K] {
	ctx, cancel := context.WithCancel(context.Background())
	task.Ctx = ctx
	task.cancel = cancel
	task.state = PENDING
	return task
}
