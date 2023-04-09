package auth

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"goapi-template/models"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/lestrrat-go/jwx/jwk"
)

func isTokenValid(r *http.Request, authConfig *models.AuthConfiguration) error {
	token, err := verifyToken(r, authConfig)
	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("token is invalid")
	}

	if token.Claims.(jwt.MapClaims).VerifyAudience(authConfig.Audience, true) {
		return fmt.Errorf("token not issue to correct audience")
	}

	return nil
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

func getJwks(jwksURL, kid string) (*rsa.PublicKey, error) {
	set, err := jwk.Fetch(context.Background(), jwksURL)
	if err != nil {
		return nil, fmt.Errorf("could not download keys")
	}

	key, ok := set.LookupKeyID(kid)
	if !ok {
		return nil, fmt.Errorf("key %v not found", kid)
	}

	publicKey := &rsa.PublicKey{}
	err = key.Raw(publicKey)
	if err != nil {
		return nil, fmt.Errorf("could not parse pubkey")
	}

	return publicKey, nil
}

func verifyToken(r *http.Request, config *models.AuthConfiguration) (*jwt.Token, error) {
	tokenString, err := extractToken(r)
	if err != nil {
		return nil, err
	}

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

		jwks, err := getJwks(config.JWKSUri, kid)
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
