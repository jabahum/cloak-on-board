package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestJWTAuthValidatesAudienceAndBuildsUser(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	provider := &JWKSProvider{
		cached:    JWKS{Keys: []JWK{jwkForKey("key", &privateKey.PublicKey)}},
		expiresAt: time.Now().Add(time.Hour),
	}
	claims := KeycloakClaims{
		PreferredUsername: "manager", Email: "manager@example.org",
		AuthorizedParty: "keycloak-onboarder-ui",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "subject", Issuer: "https://issuer/realms/onboarder",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	}
	claims.RealmAccess.Roles = []string{"manager"}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "key"
	signed, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(JWTAuth(provider, "https://issuer/realms/onboarder", "keycloak-onboarder-ui"))
	router.GET("/", func(c *gin.Context) { user, _ := GetUser(c); c.JSON(200, gin.H{"role": user.EffectiveRole()}) })
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer "+signed)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != 200 {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}

	routerWrong := gin.New()
	routerWrong.Use(JWTAuth(provider, "https://issuer/realms/onboarder", "other-client"))
	routerWrong.GET("/", func(c *gin.Context) { c.Status(204) })
	recorder = httptest.NewRecorder()
	routerWrong.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("wrong audience status=%d", recorder.Code)
	}
}

func TestJWTAuthRejectsMissingToken(t *testing.T) {
	router := gin.New()
	router.Use(JWTAuth(&JWKSProvider{}, "issuer", "audience"))
	router.GET("/", func(c *gin.Context) { c.Status(204) })
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", recorder.Code)
	}
}

func jwkForKey(kid string, key *rsa.PublicKey) JWK {
	e := big.NewInt(int64(key.E)).Bytes()
	return JWK{Kid: kid, Kty: "RSA", Alg: "RS256", Use: "sig", N: base64.RawURLEncoding.EncodeToString(key.N.Bytes()), E: base64.RawURLEncoding.EncodeToString(e)}
}
