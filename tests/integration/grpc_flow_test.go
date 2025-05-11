package integration

import (
    "context"
    "net"
    "testing"
    "time"

    proto "github.com/egocentri/finalproject/cmd/orchestrator/proto"
    grpcSrv "github.com/egocentri/finalproject/internal/grpc"
    "github.com/egocentri/finalproject/internal/config"
    "github.com/egocentri/finalproject/internal/models"
    "github.com/stretchr/testify/require"
    "google.golang.org/grpc"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func TestGRPCFlow(t *testing.T) {
    // 1) Настройка DB
    db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    db.AutoMigrate(&models.Expression{})
    db.Create(&models.Expression{Expression: "2+3", Result: ""})

    // 2) Старт gRPC-сервера на случайном порту
    cfg := config.InitEnv()
    lis, _ := net.Listen("tcp", ":0")
    srv := grpc.NewServer()
    proto.RegisterDispatcherServer(srv, grpcSrv.NewServer(db, cfg))
    go srv.Serve(lis)
    defer srv.Stop()

    // 3) Подключаемся к серверу
    conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure(), grpc.WithBlock())
    require.NoError(t, err)
    defer conn.Close()
    client := proto.NewDispatcherClient(conn)

    // 4) GetTask
    tr, err := client.GetTask(context.Background(), &proto.Empty{})
    require.NoError(t, err)
    require.Equal(t, uint32(1), tr.Task.Id)

    // 5) PostTaskResult
    ack, err := client.PostTaskResult(context.Background(), &proto.TaskResult{Id: 1, Result: "5"})
    require.NoError(t, err)
    require.True(t, ack.Ok)

    // Даем время серверу записать в БД
    time.Sleep(100 * time.Millisecond)

    var updated models.Expression
    db.First(&updated, 1)
    require.Equal(t, "5", updated.Result)
}
