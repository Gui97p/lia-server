package providers

type ProviderName string

const (
	ProviderGroq   ProviderName = "groq"
	ProviderGemini ProviderName = "gemini"
)

type Providers map[ProviderName]string

func (p ProviderName) Valid() bool {
	switch p {
	case ProviderGroq, ProviderGemini:
		return true
	default:
		return false
	}
}
