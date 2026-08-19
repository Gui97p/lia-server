package transport

import (
	"context"

	"github.com/Gui97p/lia-server/internal/users"
)

func userFromContext(ctx context.Context) (user *users.User, ok bool) {
	user, ok = ctx.Value(userCtxKey).(*users.User)
	return
}
