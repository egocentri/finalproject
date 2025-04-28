package handlers

import (
    "fmt"
    "net/http"
    "time"

    "github.com/egocentri/finalproject/internal/models"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type TasksHandler struct{ db *gorm.DB }

func NewTasksHandler(db *gorm.DB) *TasksHandler {
    return &TasksHandler{db: db}
}

// GET /internal/task — до 30 секунд
func (h *TasksHandler) GetTask(c *gin.Context) {
    timeout := time.After(30 * time.Second)
    ticker  := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-timeout:
            c.JSON(http.StatusNotFound, gin.H{"error":"no task"})
            return
        case <-ticker.C:
            var t models.Expression
            if err := h.db.Where("result = ''").First(&t).Error; err == nil {
                c.JSON(http.StatusOK, gin.H{"task": gin.H{
                    "id":                t.ID,
                    "expression":        t.Expression,
                    "operation_time_ms": 100,
                }})
                return
            }
        }
    }
}

// POST /internal/task — {id, result}
func (h *TasksHandler) PostResult(c *gin.Context) {
    var req struct{ ID uint; Result float64 }
    if c.BindJSON(&req) != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error":"bad payload"})
        return
    }
    if err := h.db.Model(&models.Expression{}).
        Where("id = ?", req.ID).
        Update("result", fmt.Sprint(req.Result)).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error":"not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error":"db error"})
        }
        return
    }
    c.JSON(http.StatusOK, gin.H{"status":"ok"})
}
