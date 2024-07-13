package auth

import (
	"fmt"
	"goapi-template/config"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
)

type JKWS interface {
	Keyfunc(token *jwt.Token) (interface{}, error)
}

func sliceContains[K comparable](s []K, e K) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func validateUserToken(tokenString string, authConfig *config.AuthConfiguration, jwks JKWS) (*User, error) {
	token, err := verifyToken(tokenString, authConfig, jwks)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	claims := token.Claims.(jwt.MapClaims)
	if claims["aud"].(string) != authConfig.Audience {
		return nil, fmt.Errorf("token not issue to correct audience")
	}

	scope := claims[authConfig.ScopeClaim].(string)

	scopeIsValid := len(authConfig.Scopes) == 0 || sliceContains(authConfig.Scopes, scope)
	if !scopeIsValid {
		return nil, fmt.Errorf("token doesn't have valid scopes")
	}

	user := &User{
		ID:     claims["sub"].(string),
		Name:   claims["name"].(string),
		Email:  claims["email"].(string),
		Claims: map[string]string{},
	}

	for _, v := range authConfig.ClaimFields {
		user.Claims[v] = claims[v].(string)
	}

	return user, nil
}

func extractToken(r *http.Request) (string, error) {
	bearerToken := r.Header.Get("Authorization")

	strArr := strings.Split(bearerToken, " ")
	if len(strArr) != 2 || strArr[1] == "" {
		return "", fmt.Errorf("invalid bearer token")
	}

	return strArr[1], nil
}

func verifyToken(tokenString string, config *config.AuthConfiguration, jwks JKWS) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, jwks.Keyfunc)

	if err != nil {
		return nil, err
	}

	algIsValid := sliceContains(config.TokenSigningAlg, token.Method.Alg())
	if !algIsValid {
		return nil, fmt.Errorf("token signature alg and issuer alg do not match")
	}

	return token, nil
}
