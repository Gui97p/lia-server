package auth

type TrustLevel string

const (
	Anonymous     TrustLevel = "anonymous"
	Identified    TrustLevel = "identified"
	Authenticated TrustLevel = "authenticated"
	Trusted       TrustLevel = "trusted"
)
