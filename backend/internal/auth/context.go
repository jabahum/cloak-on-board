package auth

import "github.com/gin-gonic/gin"

const UserContextKey = "auth_user"

type User struct {
	Subject     string
	Username    string
	Email       string
	DisplayName string
	RealmRoles  []string
}

func (u User) EffectiveRole() string {
	for _, wanted := range []string{"admin", "manager", "viewer"} {
		for _, role := range u.RealmRoles {
			if role == wanted {
				return wanted
			}
		}
	}
	return ""
}

func SetUser(c *gin.Context, user User) {
	c.Set(UserContextKey, user)
}

func GetUser(c *gin.Context) (User, bool) {
	value, exists := c.Get(UserContextKey)
	if !exists {
		return User{}, false
	}

	user, ok := value.(User)
	return user, ok
}
