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
    // Загружаем конфиг (порт gRPC, секрет и т.п.)
    cfg := config.InitEnv()

    // Подключаемся к gRPC-серверу
    addr := fmt.Sprintf("localhost:%s", cfg.GRPCPort)
    conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
    if err != nil {
        log.Fatalf("gRPC dial failed: %v", err)
    }
    defer conn.Close()

    client := proto.NewDispatcherClient(conn)

    // Основной цикл: запрашиваем задачи, вычисляем, отдаем результат
    for {
        // 1) Получаем задачу
        resp, err := client.GetTask(context.Background(), &proto.Empty{})
        if err != nil {
            // Если задач нет — ждём немного и пробуем снова
            if status.Code(err) == codes.NotFound {
                time.Sleep(500 * time.Millisecond)
                continue
            }
            // Иная ошибка — логируем и тоже ждём
            log.Println("GetTask error:", err)
            time.Sleep(time.Second)
            continue
        }

        task := resp.GetTask()
        log.Printf("Received task: ID=%d, expr=%q", task.Id, task.Expression)

        // 2) Симулируем задержку вычисления
        time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

        // 3) Вычисляем выражение
        result, err := services.Evaluate(task.Expression)
        if err != nil {
            log.Println("Evaluate error:", err)
            continue
        }

        // 4) Отправляем результат обратно
        ack, err := client.PostTaskResult(context.Background(), &proto.TaskResult{
            Id:     task.Id,
            Result: fmt.Sprint(result),
        })
        if err != nil {
            log.Println("PostTaskResult error:", err)
            continue
        }
        log.Printf("Posted result for task %d, ack: %v", task.Id, ack.Ok)
    }
}
