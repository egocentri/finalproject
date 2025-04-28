package handlers

import (
    "net/http"
    "time"

    "github.com/egocentri/finalproject/internal/models"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v4"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
)

type AuthHandler struct {
    db        *gorm.DB
    jwtSecret []byte
}

func NewAuthHandler(db *gorm.DB, secret []byte) *AuthHandler {
    return &AuthHandler{db: db, jwtSecret: secret}
}

func (h *AuthHandler) Register(c *gin.Context) {
    var req struct{ Login, Password string }
    if c.BindJSON(&req) != nil || req.Login == "" || req.Password == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "login & password required"})
        return
    }
    if err := h.db.Where("login = ?", req.Login).First(&models.User{}).Error; err == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "user exists"})
        return
    }
    hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    user := models.User{Login: req.Login, PasswordHash: string(hash)}
    if err := h.db.Create(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
        return
    }
    c.Status(http.StatusCreated)
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req struct{ Login, Password string }
    if c.BindJSON(&req) != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
        return
    }
    var user models.User
    if err := h.db.Where("login = ?", req.Login).First(&user).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
        return
    }
    if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
        return
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": user.ID,
        "exp":     time.Now().Add(24 * time.Hour).Unix(),
    })
    ts, err := token.SignedString(h.jwtSecret)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "token error"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"token": ts})
}
