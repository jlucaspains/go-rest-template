package authz

import future.keywords.if

default allow = false

allow if {
	input.token == "pass"
}