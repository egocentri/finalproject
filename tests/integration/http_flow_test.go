package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/egocentri/finalproject/internal/models"
    "github.com/egocentri/finalproject/internal/handlers"
    "github.com/egocentri/finalproject/internal/middleware"
    "github.com/egocentri/finalproject/internal/services"
    "github.com/gin-gonic/gin"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func setupServer() *gin.Engine {
    db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    db.AutoMigrate(&models.User{}, &models.Expression{})

    authH := handlers.NewAuthHandler(db, []byte("test"))
    exprH := handlers.NewExpressionsHandler(db)
    taskH := handlers.NewTasksHandler(db)

    r := gin.Default()
    api := r.Group("/api/v1")
    api.POST("/register", authH.Register)
    api.POST("/login",    authH.Login)

    sec := api.Group("/")
    sec.Use(middleware.JWTAuthMiddleware([]byte("test")))
    {
        sec.POST("/calculate",       exprH.Calculate)
        sec.GET ("/expressions",     exprH.List)
        sec.GET ("/expressions/:id", exprH.GetByID)
        sec.GET ("/internal/task",   taskH.GetTask)
        sec.POST("/internal/task",   taskH.PostResult)
    }
    return r
}

func TestFullHTTPFlow(t *testing.T) {
    r := setupServer()
    ts := httptest.NewServer(r)
    defer ts.Close()

    // 1. Register
    regBody, _ := json.Marshal(map[string]string{"login":"i","password":"p"})
    res, _ := http.Post(ts.URL+"/api/v1/register","application/json",bytes.NewBuffer(regBody))
    if res.StatusCode != http.StatusCreated { t.Fatalf("reg: %d", res.StatusCode) }

    // 2. Login
    res, _ = http.Post(ts.URL+"/api/v1/login","application/json",bytes.NewBuffer(regBody))
    var loginResp map[string]string
    json.NewDecoder(res.Body).Decode(&loginResp)
    token := loginResp["token"]
    if token == "" { t.Fatal("no token") }

    client := &http.Client{}
    auth := "Bearer " + token

    // 3. Calculate
    calcBody, _ := json.Marshal(map[string]string{"expression":"2+2"})
    req, _ := http.NewRequest("POST", ts.URL+"/api/v1/calculate", bytes.NewBuffer(calcBody))
    req.Header.Set("Content-Type","application/json")
    req.Header.Set("Authorization", auth)
    res, _ = client.Do(req)
    var calcResp map[string]interface{}
    json.NewDecoder(res.Body).Decode(&calcResp)
    if calcResp["result"] != "4" { t.Fatalf("calc: %v", calcResp) }

    // 4. List
    req, _ = http.NewRequest("GET", ts.URL+"/api/v1/expressions", nil)
    req.Header.Set("Authorization", auth)
    res, _ = client.Do(req)
    var list []map[string]interface{}
    json.NewDecoder(res.Body).Decode(&list)
    if len(list) != 1 { t.Fatalf("list len=%d", len(list)) }

}
