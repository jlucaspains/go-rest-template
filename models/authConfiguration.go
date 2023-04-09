package models

type AuthConfiguration struct {
	Issuer          string        `json:"issuer"`
	JWKSUri         string        `json:"jwks_uri"`
	TokenSigningAlg []interface{} `json:"id_token_signing_alg_values_supported"`
	Audience        string        `json:"audience"`
}
