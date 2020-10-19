package taskmanager

import (
	"fmt"
	"testing"
)

func TestTask_Exec(t1 *testing.T) {
	type fields struct {
		priority Priority
		handler  TaskHandler
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "bad",
			fields: fields{
				priority: HighestPriority,
				handler:  func() error { return fmt.Errorf("oops") },
			},
			wantErr: true,
		},
		{
			name: "good",
			fields: fields{
				priority: HighestPriority,
				handler:  func() error { return nil },
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t1.Run(tt.name, func(t1 *testing.T) {
			t := NewTask(tt.fields.priority, tt.fields.handler)
			if err := t.Exec(); (err != nil) != tt.wantErr {
				t1.Errorf("Exec() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
