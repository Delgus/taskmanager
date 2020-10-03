package taskmanager

import (
	"fmt"
	"testing"
)

func TestErrorTask(t *testing.T) {
	failFlag := false

	task := NewTask(HighestPriority, func() error { return fmt.Errorf(`test error`) })
	task.OnEvent(FailedEvent, func() {
		failFlag = true
	})
	err := task.Exec()

	if err == nil {
		t.Errorf(`expected error!`)
	}

	if !failFlag {
		t.Errorf(`expected execution of handler for failed event!`)
	}
}

func TestOnEvent(t *testing.T) {
	task := NewTask(HighestPriority, func() error { return nil })

	eventFlag := false
	task.OnEvent(BeforeExecEvent, func() {
		eventFlag = true
	})

	task.EmitEvent(AfterExecEvent)

	if eventFlag {
		t.Errorf(`unexpected execution of handler`)
	}

	task.EmitEvent(BeforeExecEvent)

	if !eventFlag {
		t.Errorf(`handler not execute`)
	}
}
