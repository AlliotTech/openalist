package task

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func waitForTaskState[K comparable](t *testing.T, task *Task[K], expected string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if task.GetState() == expected {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("task status did not become %s, got %s", expected, task.GetState())
}

func TestTask_Manager(t *testing.T) {
	tm := NewTaskManager(3, func(id *uint64) {
		atomic.AddUint64(id, 1)
	})
	id := tm.Submit(WithCancelCtx(&Task[uint64]{
		Name: "test",
		Func: func(task *Task[uint64]) error {
			time.Sleep(time.Millisecond * 500)
			return nil
		},
	}))
	task, ok := tm.Get(id)
	if !ok {
		t.Fatal("task not found")
	}
	waitForTaskState(t, task, RUNNING)
	waitForTaskState(t, task, SUCCEEDED)
}

func TestTask_Cancel(t *testing.T) {
	tm := NewTaskManager(3, func(id *uint64) {
		atomic.AddUint64(id, 1)
	})
	id := tm.Submit(WithCancelCtx(&Task[uint64]{
		Name: "test",
		Func: func(task *Task[uint64]) error {
			<-task.Ctx.Done()
			return nil
		},
	}))
	task, ok := tm.Get(id)
	if !ok {
		t.Fatal("task not found")
	}
	waitForTaskState(t, task, RUNNING)
	if err := tm.Cancel(id); err != nil {
		t.Fatal(err)
	}
	waitForTaskState(t, task, CANCELED)
}

func TestTask_Retry(t *testing.T) {
	tm := NewTaskManager(3, func(id *uint64) {
		atomic.AddUint64(id, 1)
	})
	var num atomic.Int32
	id := tm.Submit(WithCancelCtx(&Task[uint64]{
		Name: "test",
		Func: func(task *Task[uint64]) error {
			attempt := num.Add(1)
			if attempt&1 == 1 {
				return errors.New("test error")
			}
			return nil
		},
	}))
	task, ok := tm.Get(id)
	if !ok {
		t.Fatal("task not found")
	}
	waitForTaskState(t, task, ERRORED)
	if task.GetError() == nil {
		t.Error(task.GetState())
		t.Fatal("task error is nil, but expected error")
	} else {
		t.Logf("task error: %s", task.GetErrMsg())
	}
	if err := tm.Retry(id); err != nil {
		t.Fatal(err)
	}
	waitForTaskState(t, task, SUCCEEDED)
	if task.GetError() != nil {
		t.Errorf("task error: %+v, but expected nil", task.GetError())
	}
}
