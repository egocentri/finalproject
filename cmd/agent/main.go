package main

import (
    "context"
    "fmt"
    "log"
    "time"

    proto "github.com/egocentri/finalproject/cmd/orchestrator/proto"
    "github.com/egocentri/finalproject/internal/config"
    "github.com/egocentri/finalproject/internal/services"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
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
    resp, err := client.GetTask(context.Background(), &proto.Empty{})
        if err != nil {
            // Если нет задач — NotFound, ждём и повторяем
            if status.Code(err) == codes.NotFound {
                time.Sleep(500 * time.Millisecond)
                continue
            }
            // Иные ошибки выводим в лог и ждём
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
