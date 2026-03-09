package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const CtxUserID = "user_id"

func JWT(secret []byte, accesscookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenStr string

		if a := c.GetHeader("Authorization"); strings.HasPrefix(a, "Bearer ") {
			tokenStr = strings.TrimPrefix(a, "Bearer ")
		} else if v, err := c.Cookie(accesscookieName); err == nil {
			tokenStr = v
		}

		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("bad alg")
			}
			return secret, nil
		})

		if err != nil || !tok.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := tok.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "bad claims"})
			return
		}

		f, ok := claims["user_id"].(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "bad claims"})
			return
		}
		c.Set(CtxUserID, int(f))
		c.Next()
	}
}
