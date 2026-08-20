package memories

type MemoryScope string

const (
	User    MemoryScope = "user"
	Global  MemoryScope = "global"
	Private MemoryScope = "private"
)
