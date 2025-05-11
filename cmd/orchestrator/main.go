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
    "github.com/glebarez/sqlite"
    "gorm.io/gorm"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
)

func main() {
    // Загрузка конфигурации из окружения
    cfg := config.InitEnv()

    // Инициализация SQLite (чисто-Go драйвер)
    db, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
    if err != nil {
        log.Fatalf("failed to open DB: %v", err)
    }

    // Миграция схемы: создаём таблицы users и expressions
    if err := db.AutoMigrate(&models.User{}, &models.Expression{}); err != nil {
        log.Fatalf("failed to migrate DB: %v", err)
    }

    // --- HTTP API (Gin) ---
    r := gin.Default()

    // Регистрация и логин
    authH := handlers.NewAuthHandler(db, cfg.JWTSecret)
    r.POST("/api/v1/register", authH.Register)
    r.POST("/api/v1/login", authH.Login)

    // Защищённые маршруты
    exprH := handlers.NewExpressionsHandler(db)
    taskH := handlers.NewTasksHandler(db)

    sec := r.Group("/api/v1")
    sec.Use(middleware.JWTAuthMiddleware(cfg.JWTSecret))
    {
        // Пользовательские эндпоинты
        sec.POST("/calculate", exprH.Calculate)
        sec.GET("/expressions", exprH.List)
        sec.GET("/expressions/:id", exprH.GetByID)

        // Внутренние HTTP‐эндпоинты для агентов (для обратной совместимости с HTTP-тестами)
        sec.GET("/internal/task", taskH.GetTask)
        sec.POST("/internal/task", taskH.PostResult)
    }

    // --- Запуск gRPC-сервера в горутине ---
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
            log.Fatalf("gRPC serve error: %v", err)
        }
    }()

    // --- Запуск HTTP-сервера ---
    log.Printf("HTTP server listening on :%s", cfg.HTTPPort)
    if err := r.Run(":" + cfg.HTTPPort); err != nil {
        log.Fatalf("HTTP serve error: %v", err)
    }
}
