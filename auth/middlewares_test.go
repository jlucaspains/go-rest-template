package auth

import (
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	godotenv.Load("../.testing.env")

	Init()

	assert.Equal(t, "api://c571ab3c-0fde-43b2-b010-77e7bdd0d6f7/api/", authConfig.Audience)
	assert.Equal(t, "https://login.microsoftonline.com/9e6b9f31-c202-4cbd-a9b1-7e5cb3874384/v2.0", authConfig.Issuer)
	assert.Equal(t, "https://login.microsoftonline.com/9e6b9f31-c202-4cbd-a9b1-7e5cb3874384/discovery/v2.0/keys", authConfig.JWKSUri)
	assert.Equal(t, "scp", authConfig.ScopeClaim)
	assert.Equal(t, "RS256", authConfig.TokenSigningAlg[0])
	assert.NotNil(t, opaQuery)
	assert.NotNil(t, cachedSet)
}
