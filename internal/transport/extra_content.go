package transport

import (
	"context"
	"fmt"
	"strings"

	"github.com/Gui97p/lia-server/internal/memories"
	"github.com/Gui97p/lia-server/internal/session"
	"github.com/google/uuid"
)

func (s *Server) buildExtraContext(ctx context.Context, sess *session.Session, taskID uuid.UUID) (string, error) {
	summary, err := s.createSummary(ctx, sess, taskID)
	if err != nil {
		return "", err
	}

	memoryContext, err := s.createMemoryContext(ctx, sess)
	if err != nil {
		return "", err
	}

	return joinSections(summary, memoryContext), nil
}

func joinSections(sections ...string) string {
	nonEmpty := make([]string, 0, len(sections))
	for _, section := range sections {
		if len(section) > 0 {
			nonEmpty = append(nonEmpty, section)
		}
	}
	return strings.Join(nonEmpty, "\n\n")
}

func (s *Server) createSummary(ctx context.Context, sess *session.Session, taskID uuid.UUID) (string, error) {
	summary := ""
	otherTasks, err := s.tasksStore.ListByUser(ctx, sess.UserID, 5)
	if err != nil {
		return summary, err
	}
	for _, t := range otherTasks {
		if t.ID != taskID && !t.State.IsTerminal() {
			firstMsg, err := s.messagesStore.GetFirstByTask(ctx, t.ID)
			if err != nil {
				s.logger.Warn("error on get running task first message", "error", err, "task_id", t.ID)
				continue
			}
			summary += fmt.Sprintf("- %s (status: %s)\n", firstMsg.Content, t.State)
		}
	}
	if len(summary) == 0 {
		return "", nil
	}
	return "Outras tarefas em andamento nesse momento:\n" + summary, nil
}

func (s *Server) createMemoryContext(ctx context.Context, sess *session.Session) (string, error) {
	memoryContext := ""

	globalMemories, err := s.memoriesStore.ListByScope(ctx, memories.Global, 50)
	if err != nil {
		return memoryContext, err
	}

	if len(globalMemories) != 0 {
		memoryContext += "Informações gerais que você possui:\n"
		for _, m := range globalMemories {
			category := ""
			if m.Category != nil {
				category = fmt.Sprintf("[%s]", *m.Category)
			}
			memoryContext += fmt.Sprintf("- (ID: %s) [%s][%s]%s - %s\n", m.ID, m.CreatedAt.Format("2006-01-02"), m.Scope, category, m.Fact)
		}
	}

	userMemories, err := s.memoriesStore.ListByUser(ctx, sess.UserID, 50)
	if err != nil {
		return memoryContext, err
	}

	if len(userMemories) != 0 {
		memoryContext += "Memórias do usuário que você possui:\n"
		for _, m := range userMemories {
			category := ""
			if m.Category != nil {
				category = fmt.Sprintf("[%s]", *m.Category)
			}
			memoryContext += fmt.Sprintf("- (ID: %s) [%s][%s]%s - %s\n", m.ID, m.CreatedAt.Format("2006-01-02"), m.Scope, category, m.Fact)
		}
	}

	return memoryContext, nil
}
