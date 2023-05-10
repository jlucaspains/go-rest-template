package models

type AuthConfiguration struct {
	Issuer          string   `json:"issuer"`
	JWKSUri         string   `json:"jwks_uri"`
	TokenSigningAlg []string `json:"id_token_signing_alg_values_supported"`
	Audience        string   `json:"audience"`
	ScopeClaim      string   `json:"scope_claim"`
	Scopes          []string `json:"scopes"`
	ClaimFields     []string `json:"claims"`
}
