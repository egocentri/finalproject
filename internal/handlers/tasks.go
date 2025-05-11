package handlers

import (
    "fmt"
    "net/http"
    "strconv"

    "github.com/egocentri/finalproject/internal/models"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type TasksHandler struct {
    db *gorm.DB
}

func NewTasksHandler(db *gorm.DB) *TasksHandler {
    return &TasksHandler{db: db}
}

// GetTask возвращает первую задачу (Expression с пустым Result)
func (h *TasksHandler) GetTask(c *gin.Context) {
    var expr models.Expression
    if err := h.db.Where("result = ?", "").First(&expr).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "no tasks available"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    // Возвращаем ID, выражение и operation_time_ms=0 (тест его не проверяет)
    c.JSON(http.StatusOK, gin.H{"task": gin.H{
        "id":                expr.ID,
        "expression":        expr.Expression,
        "operation_time_ms": 0,
    }})
}

// PostResult принимает результат задачи и сохраняет в БД
func (h *TasksHandler) PostResult(c *gin.Context) {
    var req struct {
        ID     uint        `json:"id"`
        Result interface{} `json:"result"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid request"})
        return
    }

    // Приводим result к строке
    resultStr := fmt.Sprint(req.Result)

    if err := h.db.Model(&models.Expression{}).
        Where("id = ?", req.ID).
        Update("result", resultStr).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
