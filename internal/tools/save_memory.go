package tools

import (
	"context"
	"fmt"

	"github.com/Gui97p/lia-server/internal/memories"
	"github.com/Gui97p/lia-server/internal/session"
)

func NewSaveMemoryHandler(store memories.Store) Handler {
	return func(ctx context.Context, sess *session.Session, params map[string]any) (session.ToolResult, error) {
		scopeStr, ok := params["scope"].(string)
		if !ok || len(scopeStr) == 0 {
			err := fmt.Errorf("scope is needed")
			return session.ToolResult{Success: false, Error: err.Error()}, err
		}
		scope := memories.MemoryScope(scopeStr)

		fact, ok := params["fact"].(string)
		if !ok || len(fact) == 0 {
			err := fmt.Errorf("fact is needed")
			return session.ToolResult{Success: false, Error: err.Error()}, err
		}

		var memory *memories.Memory
		var err error

		switch scope {
		case memories.Global:
			memory, err = store.Create(ctx, scope, fact, nil)
		case memories.Private:
			memory, err = store.Create(ctx, scope, fact, nil)
		case memories.User:
			memory, err = store.Create(ctx, scope, fact, &sess.UserID)
		default:
			err = fmt.Errorf("invalid scope")
			return session.ToolResult{Success: false, Error: err.Error()}, err
		}
		if err != nil {
			return session.ToolResult{Success: false, Error: err.Error()}, err
		}

		category, ok := params["category"].(string)
		if ok {
			if len(category) > 0 {
				err = store.SetCategory(ctx, memory.ID, category)
				if err != nil {
					return session.ToolResult{Success: false, Error: err.Error()}, err
				}
			}
		}

		return session.ToolResult{
			Success: true,
		}, nil
	}
}
