package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/egocentri/finalproject/cmd/orchestrator/proto"
    "github.com/egocentri/finalproject/internal/config"
    "github.com/egocentri/finalproject/internal/services"
    "google.golang.org/grpc"
)

func main() {
    cfg := config.InitEnv()

    addr := fmt.Sprintf("localhost:%s", cfg.GRPCPort)
    conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
    if err != nil {
        log.Fatalf("gRPC dial: %v", err)
    }
    defer conn.Close()
    client := proto.NewDispatcherClient(conn)

    for {
        // 1) Запрос задачи
        resp, err := client.GetTask(context.Background(), &proto.Empty{})
        if err != nil {
            log.Println("GetTask error:", err)
            time.Sleep(time.Second)
            continue
        }
        task := resp.GetTask()
        log.Printf("Received task: ID=%d, expr=%s", task.Id, task.Expression)

        // 2) Ждём operation_time
        time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

        // 3) Вычисляем
        result, err := services.Evaluate(task.Expression)
        if err != nil {
            log.Println("Evaluate error:", err)
            continue
        }

        // 4) Отправляем результат
        ack, err := client.PostTaskResult(context.Background(), &proto.TaskResult{
            Id:     task.Id,
            Result: fmt.Sprint(result),
        })
        if err != nil {
            log.Println("PostTaskResult error:", err)
        } else {
            log.Printf("Posted result. Ack: %v", ack.Ok)
        }
    }
}
