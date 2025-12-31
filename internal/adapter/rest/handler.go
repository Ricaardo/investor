package rest

import (
	"context"
	"investor/internal/core"
	"investor/internal/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Adapter struct {
	Dispatcher *core.Dispatcher
	Logger     *zap.Logger
	Port       string
}

func NewAdapter(port string, dispatcher *core.Dispatcher, logger *zap.Logger) *Adapter {
	return &Adapter{
		Dispatcher: dispatcher,
		Logger:     logger,
		Port:       port,
	}
}

type ChatRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Text     string `json:"text" binding:"required"`
	ChatID   string `json:"chat_id"`
	Platform string `json:"platform"` // optional, defaults to "api"
}

type ChatResponse struct {
	Response string `json:"response"`
}

func (a *Adapter) Start(ctx context.Context) error {
	r := gin.Default()

	r.POST("/api/v1/chat", a.handleChat)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	a.Logger.Info("Starting REST API server", zap.String("port", a.Port))
	return r.Run(":" + a.Port)
}

func (a *Adapter) handleChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	platform := req.Platform
	if platform == "" {
		platform = "api"
	}

	msg := &model.InternalMessage{
		Platform:    platform,
		ChatType:    "private",
		ChatID:      req.ChatID,
		UserID:      req.UserID,
		Text:        req.Text,
		Timestamp:   time.Now().Unix(),
		IsMentioned: true, // API calls are always mentions/direct
	}

	// Dispatch is currently synchronous for the API response,
	// but the dispatcher might run async.
	// We need to wait for the response to return it in HTTP.
	// The current Dispatcher.Dispatch returns (string, error), so it is synchronous.
	// If it spawns goroutines, we might need to adjust, but core.Dispatcher.Dispatch calls agent.Process which seems sync.

	respText, err := a.Dispatcher.Dispatch(c.Request.Context(), msg)
	if err != nil {
		a.Logger.Error("Dispatch failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, ChatResponse{
		Response: respText,
	})
}
