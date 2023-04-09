package authz_adv

import future.keywords.if

default allow = false

allow if {
	payload.email == "jlucaspains@gmail.com"
	payload.verified
	startswith(input.path, "/person/1")
}

# Cache response for 24 hours
metadata_discovery(issuer) := http.send({
	"url": concat("", [issuer, "/.well-known/openid-configuration"]),
	"method": "GET",
	"force_cache": true,
	"force_cache_duration_seconds": 86400,
}).body

# Cache response for 1 hour
jwks_request(url) := http.send({
	"url": url,
	"method": "GET",
	"force_cache": true,
	"force_cache_duration_seconds": 3600, # Cache response for an hour
})

payload := {"verified": verified, "email": payload.email} if {
	[headers, payload, _] := io.jwt.decode(input.token)
	metadata := metadata_discovery(payload.iss)

	jwks_endpoint := metadata.jwks_uri

	jwks := jwks_request(jwks_endpoint).raw_body

	verified := io.jwt.verify_rs256(input.token, jwks)
}
