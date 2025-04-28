package main

import (
    "log"
    "os"
    "github.com/egocentri/finalproject/internal/handlers"
    "github.com/egocentri/finalproject/internal/middleware"
    "github.com/egocentri/finalproject/internal/models"
    "github.com/gin-gonic/gin"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

var jwtSecret = []byte("supersecretkey") 
func main() {
    db, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
    if err != nil {
        log.Fatalf("DB connect error: %v", err)
    }
    if err := db.AutoMigrate(&models.User{}, &models.Expression{}); err != nil {
        log.Fatalf("DB migrate error: %v", err)
    }
    r := gin.Default()
    authH := handlers.NewAuthHandler(db, jwtSecret)
    exprH := handlers.NewExpressionsHandler(db)
    taskH := handlers.NewTasksHandler(db)

    api := r.Group("/api/v1")
    {
        api.POST("/register", authH.Register)
        api.POST("/login",    authH.Login)
    }
    apiAuth := api.Group("/")
    apiAuth.Use(middleware.JWTAuthMiddleware(jwtSecret))
    {
        apiAuth.POST("/calculate",   exprH.Calculate)
        apiAuth.GET ("/expressions", exprH.List)
        apiAuth.GET ("/expressions/:id", exprH.GetByID)
        // внутренние эндпоинты для агента
        apiAuth.GET ("/internal/task",       taskH.GetTask)
        apiAuth.POST("/internal/task",       taskH.PostResult)
    }
    port := "8080"
    if p := os.Getenv("PORT"); p != "" {
        port = p
    }
    log.Printf("HTTP server on :%s", port)
    if err := r.Run(":" + port); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
