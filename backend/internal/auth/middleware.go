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
	Name              string `json:"name"`
	AuthorizedParty   string `json:"azp"`
	RealmAccess       struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	jwt.RegisteredClaims
}

func JWTAuth(provider *JWKSProvider, issuer, audience string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenValue := strings.TrimPrefix(authHeader, "Bearer ")
		tokenValue = strings.TrimSpace(tokenValue)

		if tokenValue == "" {
			unauthorized(c, "missing bearer token")
			return
		}

		claims := &KeycloakClaims{}

		refreshed := false
		token, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
				return nil, errors.New("unsupported token algorithm")
			}
			kid, ok := token.Header["kid"].(string)
			if !ok || kid == "" {
				return nil, errors.New("missing token kid")
			}

			for {
				jwks, err := provider.Get(c.Request.Context())
				if err != nil {
					return nil, err
				}
				for _, key := range jwks.Keys {
					if key.Kid == kid {
						return rsaPublicKeyFromJWK(key)
					}
				}
				if refreshed {
					break
				}
				provider.Invalidate()
				refreshed = true
			}
			return nil, errors.New("matching jwk not found")
		}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}), jwt.WithIssuer(issuer))
		if err != nil || !token.Valid {
			unauthorized(c, "invalid bearer token")
			return
		}

		if audience != "" && !contains(claims.Audience, audience) && claims.AuthorizedParty != audience {
			unauthorized(c, "invalid token audience")
			return
		}

		subject := claims.Subject
		if subject == "" {
			subject = "username:" + claims.PreferredUsername
		}
		SetUser(c, User{
			Subject:     subject,
			Username:    claims.PreferredUsername,
			Email:       claims.Email,
			DisplayName: claims.Name,
			RealmRoles:  claims.RealmAccess.Roles,
		})
		user, _ := GetUser(c)
		if user.EffectiveRole() == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":      "no supported cloak-on-board role",
				"request_id": requestID(c),
			})
			return
		}

		c.Next()
	}
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func unauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": message, "request_id": requestID(c),
	})
}

func requestID(c *gin.Context) string {
	value, _ := c.Get("request_id")
	id, _ := value.(string)
	return id
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
