package core

import (
	"testing"
)

func TestNewCleanupTask(t *testing.T) {
	task := NewCleanupTask()

	if task.Interval != 30 {
		t.Errorf("Expected interval to be 30, but was %d", task.Interval)
	}
	if task.MaxImages != 900 {
		t.Errorf("Expected max images to be 900, but was %d", task.MaxImages)
	}
	if task.AwsRegion != "us-east-1" {
		t.Errorf("Expected aws region to be 'us-east-1', but was %s", task.AwsRegion)
	}
}
