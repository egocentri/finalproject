package main

import (
    "log"
    "net"
    "os"

    "github.com/egocentri/finalproject/internal/grpc"
    "github.com/egocentri/finalproject/internal/handlers"
    "github.com/egocentri/finalproject/internal/middleware"
    "github.com/egocentri/finalproject/internal/models"
    "github.com/gin-gonic/gin"
    jwt "github.com/golang-jwt/jwt/v4"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "google.golang.org/grpc"
)

var jwtSecret = []byte("supersecretkey") 
func main() {
    db, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
    if err != nil {
        log.Fatalf("failed to connect database: %v", err)
    }
    if err := db.AutoMigrate(&models.User{}, &models.Expression{}); err != nil {
        log.Fatalf("migration failed: %v", err)
    }

    r := gin.Default()
    authH := handlers.NewAuthHandler(db, jwtSecret)
    exprH := handlers.NewExpressionsHandler(db)

    api := r.Group("/api/v1")
    api.POST("/register", authH.Register)
    api.POST("/login", authH.Login)
    apiAuth := api.Group("/")
    apiAuth.Use(middleware.JWTAuthMiddleware(jwtSecret))
    {
        apiAuth.POST("/calculate", exprH.Calculate)
        apiAuth.GET("/expressions", exprH.List)
        apiAuth.GET("/expressions/:id", exprH.GetByID)
    }

    go func() {
        addr := ":8080"
        if p := os.Getenv("PORT"); p != "" {
            addr = ":" + p
        }
        log.Printf("HTTP server on %s", addr)
        if err := r.Run(addr); err != nil {
            log.Fatalf("HTTP run error: %v", err)
        }
    }()
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    grpcServer := grpc.NewServer()
    grpc.RegisterDispatcherServer(grpcServer, grpc.NewServerImpl(db))
    log.Println("gRPC server on :50051")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("gRPC serve error: %v", err)
    }
}
