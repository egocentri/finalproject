package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    jwt "github.com/golang-jwt/jwt/v4"
)

const UserIDKey = "userID"

func JWTAuthMiddleware(secret []byte) gin.HandlerFunc {
    return func(c *gin.Context) {
        h := c.GetHeader("Authorization")
        if h == "" || !strings.HasPrefix(h, "Bearer ") {
            c.AbortWithStatus(http.StatusUnauthorized)
            return
        }
        tok := strings.TrimPrefix(h, "Bearer ")
        token, err := jwt.Parse(tok, func(t *jwt.Token) (interface{}, error) {
            if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrSignatureInvalid
            }
            return secret, nil
        })
        if err != nil || !token.Valid {
            c.AbortWithStatus(http.StatusUnauthorized)
            return
        }
        claims := token.Claims.(jwt.MapClaims)
        uid := uint(claims["user_id"].(float64))
        c.Set(UserIDKey, uid)
        c.Next()
    }
}
