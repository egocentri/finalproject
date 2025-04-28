package main

import (
    "log"
    "os"
    "time"
    "bytes"
    "github.com/egocentri/finalproject/internal/services"
    "net/http"
    "encoding/json"
)

func main() {
    orchestratorURL := "http://localhost:8080"
    if u := os.Getenv("ORCHESTRATOR_URL"); u != "" {
        orchestratorURL = u
    }
    for {
        // 1) Получаем задачу
        resp, err := http.Get(orchestratorURL + "/api/v1/internal/task")
        if err != nil {
            log.Println("GetTask error:", err)
            time.Sleep(time.Second)
            continue
        }
        if resp.StatusCode != http.StatusOK {
            log.Println("No task available, status:", resp.StatusCode)
            time.Sleep(time.Second)
            continue
        }
        var body struct {
            Task struct {
                ID                uint   `json:"id"`
                Expression        string `json:"expression"`
                OperationTimeMs   int    `json:"operation_time_ms"`
            } `json:"task"`
        }
        json.NewDecoder(resp.Body).Decode(&body)
        resp.Body.Close()
        task := body.Task
        log.Printf("Got task %d: %s", task.ID, task.Expression)

        // 2) Ждём operation_time_ms
        time.Sleep(time.Duration(task.OperationTimeMs) * time.Millisecond)

        // 3) Вычисляем
        result, err := services.Evaluate(task.Expression)
        if err != nil {
            log.Println("Evaluate error:", err)
            continue
        }

        // 4) Отправляем результат
        payload, _ := json.Marshal(map[string]interface{}{
            "id":     task.ID,
            "result": result,
        })
        postResp, err := http.Post(orchestratorURL + "/api/v1/internal/task",
            "application/json", bytes.NewBuffer(payload))
        if err != nil {
            log.Println("PostResult error:", err)
            continue
        }
        postResp.Body.Close()
        log.Printf("Submitted result for %d → %v (status %d)", task.ID, result, postResp.StatusCode)
    }
}
