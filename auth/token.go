package auth

import (
	"encoding/json"
	"fmt"
	"goapi-template/models"
	"goapi-template/util"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
)

type JKWS interface {
	Keyfunc(token *jwt.Token) (interface{}, error)
}

func validateUserToken(tokenString string, authConfig *models.AuthConfiguration, jwks JKWS) (*User, error) {
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

	scopeIsValid := len(authConfig.Scopes) == 0 || util.Contains(authConfig.Scopes, scope)
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

func readAuthConfiguration(configUrl string) (*models.AuthConfiguration, error) {
	if configUrl == "" {
		return nil, fmt.Errorf("cannot read OpenId configuration without URL")
	}

	response, err := http.Get(configUrl)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	target := new(models.AuthConfiguration)
	err = json.NewDecoder(response.Body).Decode(target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

func extractToken(r *http.Request) (string, error) {
	bearerToken := r.Header.Get("Authorization")

	strArr := strings.Split(bearerToken, " ")
	if len(strArr) != 2 {
		return "", fmt.Errorf("invalid bearer token")
	}

	return strArr[1], nil
}

func verifyToken(tokenString string, config *models.AuthConfiguration, jwks JKWS) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, jwks.Keyfunc)

	if err != nil {
		return nil, err
	}

	algIsValid := util.Contains(config.TokenSigningAlg, token.Method.Alg())
	if !algIsValid {
		return nil, fmt.Errorf("token signature alg and issuer alg do not match")
	}

	return token, nil
}
