package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"math/big"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type KeycloakClaims struct {
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	RealmAccess       struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	jwt.RegisteredClaims
}

func JWTAuth(provider *JWKSProvider, issuer string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenValue := strings.TrimPrefix(authHeader, "Bearer ")
		tokenValue = strings.TrimSpace(tokenValue)

		if tokenValue == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing bearer token",
			})
			return
		}

		claims := &KeycloakClaims{}

		token, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (any, error) {
			kid, ok := token.Header["kid"].(string)
			if !ok || kid == "" {
				return nil, errors.New("missing token kid")
			}

			jwks, err := provider.Get(c.Request.Context())
			if err != nil {
				return nil, err
			}

			for _, key := range jwks.Keys {
				if key.Kid == kid {
					return rsaPublicKeyFromJWK(key)
				}
			}

			return nil, errors.New("matching jwk not found")
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid bearer token",
			})
			return
		}

		if claims.Issuer != issuer {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token issuer",
			})
			return
		}

		SetUser(c, User{
			Subject:    claims.Subject,
			Username:   claims.PreferredUsername,
			Email:      claims.Email,
			RealmRoles: claims.RealmAccess.Roles,
		})

		c.Next()
	}
}

func rsaPublicKeyFromJWK(key JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, err
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes).Int64()

	return &rsa.PublicKey{
		N: n,
		E: int(e),
	}, nil
}
