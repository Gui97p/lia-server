package tasks

type TaskState string

const (
	Created    TaskState = "created"
	Planning   TaskState = "planning"
	Ready      TaskState = "ready"
	Running    TaskState = "running"
	Waiting    TaskState = "waiting"
	Blocked    TaskState = "blocked"
	Replanning TaskState = "replanning"
	Completed  TaskState = "completed"
	Failed     TaskState = "failed"
	Cancelled  TaskState = "cancelled"
)

type TriggerType string

const (
	User      TriggerType = "user"
	Scheduled TriggerType = "scheduled"
	Event     TriggerType = "event"
)
