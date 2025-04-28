// tests/integration/agent_flow_test.go
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/egocentri/finalproject/internal/handlers"
	"github.com/egocentri/finalproject/internal/middleware"
	"github.com/egocentri/finalproject/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupServerAndDB создает в памяти GORM-базу, настраивает Gin с маршрутами
// и возвращает сервер, DB и секрет для JWT.
func setupServerAndDB(t *testing.T) (*gin.Engine, *gorm.DB, []byte) {
	t.Helper()

	// In-memory SQLite
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory DB: %v", err)
	}
	// Миграции
	if err := db.AutoMigrate(&models.User{}, &models.Expression{}); err != nil {
		t.Fatalf("failed migrate: %v", err)
	}

	// Секрет для тестов
	secret := []byte("testsecret")

	// Handlers
	authH := handlers.NewAuthHandler(db, secret)
	exprH := handlers.NewExpressionsHandler(db)
	taskH := handlers.NewTasksHandler(db)

	// Настройка Gin
	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		v1.POST("/register", authH.Register)
		v1.POST("/login", authH.Login)
	}
	secured := v1.Group("/")
	secured.Use(middleware.JWTAuthMiddleware(secret))
	{
		secured.POST("/calculate", exprH.Calculate)
		secured.GET("/expressions", exprH.List)
		secured.GET("/expressions/:id", exprH.GetByID)
		secured.GET("/internal/task", taskH.GetTask)
		secured.POST("/internal/task", taskH.PostResult)
	}

	return r, db, secret
}

func TestAgentFlow(t *testing.T) {
	// Подготовка сервера и БД
	r, db, secret := setupServerAndDB(t)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Добавляем пользователя в БД напрямую
	pwHash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt error: %v", err)
	}
	user := models.User{Login: "agentuser", PasswordHash: string(pwHash)}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user error: %v", err)
	}

	// Генерируем JWT для этого пользователя
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := tokenObj.SignedString(secret)
	if err != nil {
		t.Fatalf("token sign error: %v", err)
	}
	authHeader := "Bearer " + tokenStr

	// Сидим новую задачу (Expression) с пустым Result
	expr := models.Expression{UserID: user.ID, Expression: "3*3", Result: ""}
	if err := db.Create(&expr).Error; err != nil {
		t.Fatalf("seed expression error: %v", err)
	}

	// 1) GET /api/v1/internal/task
	req, _ := http.NewRequest("GET", ts.URL+"/api/v1/internal/task", nil)
	req.Header.Set("Authorization", authHeader)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GetTask request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GetTask expected 200, got %d", resp.StatusCode)
	}

	var getResp struct {
		Task struct {
			ID               uint   `json:"id"`
			Expression       string `json:"expression"`
			OperationTimeMs  int    `json:"operation_time_ms"`
		} `json:"task"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&getResp); err != nil {
		t.Fatalf("decode GetTask response error: %v", err)
	}
	if getResp.Task.ID != expr.ID {
		t.Fatalf("GetTask returned id %d, want %d", getResp.Task.ID, expr.ID)
	}
	if getResp.Task.Expression != expr.Expression {
		t.Fatalf("GetTask returned expression %q, want %q", getResp.Task.Expression, expr.Expression)
	}
	resultPayload, _ := json.Marshal(map[string]interface{}{
		"id":     expr.ID,
		"result": 9,
	})
	postReq, _ := http.NewRequest("POST", ts.URL+"/api/v1/internal/task", bytes.NewBuffer(resultPayload))
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Authorization", authHeader)
	postResp, err := http.DefaultClient.Do(postReq)
	if err != nil {
		t.Fatalf("PostResult request error: %v", err)
	}
	defer postResp.Body.Close()
	if postResp.StatusCode != http.StatusOK {
		t.Fatalf("PostResult expected 200, got %d", postResp.StatusCode)
	}
	var postRespBody map[string]string
	if err := json.NewDecoder(postResp.Body).Decode(&postRespBody); err != nil {
		t.Fatalf("decode PostResult response error: %v", err)
	}
	if postRespBody["status"] != "ok" {
		t.Fatalf("PostResult returned status %q, want \"ok\"", postRespBody["status"])
	}

	// 3) Проверяем в БД, что Result обновился
	var updated models.Expression
	if err := db.First(&updated, expr.ID).Error; err != nil {
		t.Fatalf("DB fetch updated expr error: %v", err)
	}
	if updated.Result != "9" {
		t.Fatalf("DB stored result %q, want \"9\"", updated.Result)
	}
}
