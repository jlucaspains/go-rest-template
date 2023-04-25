package auth

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"goapi-template/models"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

func validateUserToken(tokenString string, authConfig *models.AuthConfiguration, jwks jwk.Set) (*User, error) {
	token, err := verifyToken(tokenString, authConfig, jwks)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	claims := token.Claims.(jwt.MapClaims)
	if claims.VerifyAudience(authConfig.Audience, true) {
		return nil, fmt.Errorf("token not issue to correct audience")
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

func getJwks(jwks jwk.Set, kid string) (*rsa.PublicKey, error) {
	key, ok := jwks.LookupKeyID(kid)
	if !ok {
		return nil, fmt.Errorf("key %v not found", kid)
	}

	publicKey := &rsa.PublicKey{}
	err := key.Raw(publicKey)
	if err != nil {
		return nil, fmt.Errorf("could not parse pubkey")
	}

	return publicKey, nil
}

func verifyToken(tokenString string, config *models.AuthConfiguration, jwks jwk.Set) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid header not found")
		}

		algIsValid := false
		for _, v := range config.TokenSigningAlg {
			algIsValid = v.(string) == token.Method.Alg()

			if algIsValid {
				break
			}
		}

		if !algIsValid {
			return nil, fmt.Errorf("token signature alg and issuer alg do not match")
		}

		jwks, err := getJwks(jwks, kid)
		if err != nil {
			return nil, err
		}

		return jwks, nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}
