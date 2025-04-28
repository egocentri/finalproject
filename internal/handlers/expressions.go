package handlers

import (
    "fmt"
    "net/http"
    "strconv"

    "github.com/egocentri/finalproject/internal/middleware"
    "github.com/egocentri/finalproject/internal/models"
    "github.com/egocentri/finalproject/internal/services"
    "github.com/gin-gonic/gin"
)

type ExpressionsHandler struct{ db *gorm.DB }

func NewExpressionsHandler(db *gorm.DB) *ExpressionsHandler {
    return &ExpressionsHandler{db: db}
}

func (h *ExpressionsHandler) Calculate(c *gin.Context) {
    uid := c.GetUint(middleware.UserIDKey)
    var req struct{ Expression string }
    if c.BindJSON(&req) != nil || req.Expression == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "expr required"})
        return
    }
    res, err := services.Evaluate(req.Expression)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    expr := models.Expression{
        UserID:     uid,
        Expression: req.Expression,
        Result:     fmt.Sprint(res),
    }
    if err := h.db.Create(&expr).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"id": expr.ID, "result": expr.Result})
}

func (h *ExpressionsHandler) List(c *gin.Context) {
    uid := c.GetUint(middleware.UserIDKey)
    var out []models.Expression
    h.db.Where("user_id = ?", uid).Order("created_at desc").Find(&out)
    c.JSON(http.StatusOK, out)
}

func (h *ExpressionsHandler) GetByID(c *gin.Context) {
    uid := c.GetUint(middleware.UserIDKey)
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    var expr models.Expression
    if err := h.db.Where("id = ? AND user_id = ?", id, uid).First(&expr).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    c.JSON(http.StatusOK, expr)
}
