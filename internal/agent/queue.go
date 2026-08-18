package agent

import (
	"context"
	"sync"

	"github.com/Gui97p/lia-server/internal/llm"
	"github.com/Gui97p/lia-server/internal/session"
	"github.com/google/uuid"
)

type PlanningQueueManager struct {
	mu      sync.Mutex
	queues  map[uuid.UUID]chan planJob
	planner *Planner
}

type planJob struct {
	ctx          context.Context
	apiKey       string
	history      []llm.Message
	capabilities []string
	result       chan PlanResult
}

type PlanResult struct {
	reply string
	step  *Step
	err   error
}

func (p *PlanningQueueManager) worker(ch chan planJob) {
	for job := range ch {
		reply, step, err := p.planner.Plan(job.ctx, job.apiKey, job.history, job.capabilities)
		job.result <- PlanResult{reply: reply, step: step, err: err}
	}
}

func (p *PlanningQueueManager) EnsureStarted(userID uuid.UUID) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.queues == nil {
		p.queues = make(map[uuid.UUID]chan planJob)
	}

	if p.queues[userID] == nil {
		p.queues[userID] = make(chan planJob)
		go p.worker(p.queues[userID])
	}
}

func (p *PlanningQueueManager) StopIfUnused(userID uuid.UUID, hub *session.Hub) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(hub.FindByUser(userID)) == 0 && p.queues[userID] != nil {
		close(p.queues[userID])
		delete(p.queues, userID)
	}
}

func (p *PlanningQueueManager) Submit(ctx context.Context, userID uuid.UUID, apiKey string, history []llm.Message, capabilities []string) *PlanResult {
	p.mu.Lock()
	ch := p.queues[userID]
	p.mu.Unlock()

	if ch == nil {
		return nil
	}

	job := planJob{
		ctx:          ctx,
		apiKey:       apiKey,
		history:      history,
		capabilities: capabilities,
		result:       make(chan PlanResult),
	}

	ch <- job
	result := <-job.result
	return &result
}
