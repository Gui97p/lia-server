package providers

type ProviderName string

const (
	ProviderGroq   ProviderName = "groq"
	ProviderGemini ProviderName = "gemini"
)

var ProviderList []string = []string{string(ProviderGroq), string(ProviderGemini)}

type Providers map[ProviderName]string

func (p ProviderName) Valid() bool {
	switch p {
	case ProviderGroq, ProviderGemini:
		return true
	default:
		return false
	}
}
