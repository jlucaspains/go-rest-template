package authz

import future.keywords.if

default allow = false

allow if {
	endswith(payload.email, "@gmail.com")
	payload.verified
	startswith(input.path, "/person")
}

payload := {"verified": verified, "email": payload.email} if {
	[_, payload, _] := io.jwt.decode(input.token)
	verified := true
}
