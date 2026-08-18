package agent

type Step struct {
	ID         string
	Capability string
	Params     map[string]any
	DependsOn  []string
}

type Workflow struct {
	Steps []Step
}
