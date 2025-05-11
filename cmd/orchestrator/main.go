package main

import (
    "log"
    "net"

    proto "github.com/egocentri/finalproject/cmd/orchestrator/proto"
    grpcSrv "github.com/egocentri/finalproject/internal/grpc"
    "github.com/egocentri/finalproject/internal/config"
    "github.com/egocentri/finalproject/internal/handlers"
    "github.com/egocentri/finalproject/internal/middleware"
    "github.com/egocentri/finalproject/internal/models"
    "github.com/gin-gonic/gin"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
    "github.com/glebarez/sqlite"
    "gorm.io/gorm"
)

func main() {
    cfg := config.InitEnv()

    // ——— Настройка БД ———
    db, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
    if err != nil {
        log.Fatalf("failed to open DB: %v", err)
    }
    db.AutoMigrate(&models.User{}, &models.Expression{})

    // ——— HTTP API для пользователей ———
    r := gin.Default()
    authH := handlers.NewAuthHandler(db, cfg.JWTSecret)
    exprH := handlers.NewExpressionsHandler(db)

    api := r.Group("/api/v1")
    api.POST("/register", authH.Register)
    api.POST("/login",    authH.Login)

    secure := api.Group("/")
    secure.Use(middleware.JWTAuthMiddleware(cfg.JWTSecret))
    {
        secure.POST("/calculate",   exprH.Calculate)
        secure.GET ("/expressions", exprH.List)
        secure.GET ("/expressions/:id", exprH.GetByID)
    }

    // ——— gRPC-сервер для агентов ———
    go func() {
        lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
        if err != nil {
            log.Fatalf("gRPC listen failed: %v", err)
        }
        grpcServer := grpc.NewServer()
        proto.RegisterDispatcherServer(grpcServer, grpcSrv.NewServer(db, cfg))
        reflection.Register(grpcServer)
        log.Printf("gRPC server listening on :%s", cfg.GRPCPort)
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("gRPC serve: %v", err)
        }
    }()

    // ——— Запускаем HTTP ———
    log.Printf("HTTP server listening on :%s", cfg.HTTPPort)
    if err := r.Run(":" + cfg.HTTPPort); err != nil {
        log.Fatalf("HTTP serve: %v", err)
    }
}

