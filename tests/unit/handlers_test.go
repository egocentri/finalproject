package unit

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/egocentri/finalproject/internal/models"
    "github.com/egocentri/finalproject/internal/handlers"
    "github.com/gin-gonic/gin"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func setupAuth() *gin.Engine {
    db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    db.AutoMigrate(&models.User{}, &models.Expression{})
    r := gin.Default()
    h := handlers.NewAuthHandler(db, []byte("testkey"))
    v1 := r.Group("/api/v1")
    { v1.POST("/register", h.Register); v1.POST("/login", h.Login) }
    return r
}

func TestRegisterAndLogin(t *testing.T) {
    r := setupAuth()
    body, _ := json.Marshal(gin.H{"login":"u","password":"p"})
    w := httptest.NewRecorder()
    req := httptest.NewRequest("POST","/api/v1/register",bytes.NewBuffer(body))
    req.Header.Set("Content-Type","application/json")
    r.ServeHTTP(w, req)
    if w.Code != http.StatusCreated {
        t.Fatalf("expected 201, got %d", w.Code)
    }

    w = httptest.NewRecorder()
    req = httptest.NewRequest("POST","/api/v1/login",bytes.NewBuffer(body))
    req.Header.Set("Content-Type","application/json")
    r.ServeHTTP(w, req)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", w.Code)
    }
    var resp map[string]string
    json.NewDecoder(w.Body).Decode(&resp)
    if resp["token"] == "" {
        t.Fatal("expected token")
    }
}
