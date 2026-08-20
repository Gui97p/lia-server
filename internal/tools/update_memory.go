package tools

import (
	"context"
	"fmt"

	"github.com/Gui97p/lia-server/internal/memories"
	"github.com/Gui97p/lia-server/internal/session"
	"github.com/google/uuid"
)

func NewUpdateMemoryHandler(store memories.Store) Handler {
	return func(ctx context.Context, sess *session.Session, params map[string]any) (session.ToolResult, error) {
		idStr, ok := params["id"].(string)
		if !ok || len(idStr) == 0 {
			err := fmt.Errorf("id is needed")
			return session.ToolResult{Success: false, Error: err.Error()}, err
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			return session.ToolResult{Success: false, Error: err.Error()}, err
		}

		fact, ok := params["fact"].(string)
		if !ok || len(fact) == 0 {
			err := fmt.Errorf("fact is needed")
			return session.ToolResult{Success: false, Error: err.Error()}, err
		}

		err = store.SetFact(ctx, id, fact)
		if err != nil {
			return session.ToolResult{Success: false, Error: err.Error()}, err
		}

		return session.ToolResult{
			Success: true,
		}, nil
	}
}
