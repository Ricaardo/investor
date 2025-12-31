package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"investor/config"
	"investor/internal/adapter/feishu"
	"investor/internal/adapter/rest"
	"investor/internal/agent"
	"investor/internal/core"
	"investor/internal/dataservice"
	"investor/internal/llm"
	"investor/internal/session"
)

func main() {
	// 1. Init Config
	config.Init()

	// 2. Init Logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 3. Init LLM
	llmProvider := llm.NewOpenAIProvider(config.AppConfig.LLM)

	// 4. Init Core Services
	sessionMgr := session.NewManager()

	// 4.1 Init Data Service Registry (Extensible Data Sources)
	registry := dataservice.GetRegistry()
	// Register Yahoo (Primary)
	yahooSvc := dataservice.NewYahooDataService()
	registry.Register("yahoo", yahooSvc)
	// TODO: Register other data sources here (e.g., Bloomberg, Custom API)

	// Use Default Data Service for Agent
	dataService := registry.GetDefault()

	// 5. Init Agents
	// Note: We only need ChatAgent now, as it handles IPO intent too via Tools
	chatAgent := agent.NewChatAgent(llmProvider, sessionMgr, dataService)

	// 6. Init Dispatcher
	dispatcher := core.NewDispatcher(logger)
	dispatcher.RegisterAgent(chatAgent)

	// 7. Init Adapters (Multi-Channel Support)

	// 7.1 Feishu Adapter (WebSocket Mode)
	if config.AppConfig.Feishu.AppID != "" && config.AppConfig.Feishu.AppSecret != "" {
		feishuAdapter := feishu.NewAdapter(config.AppConfig.Feishu, dispatcher, logger)
		go func() {
			if err := feishuAdapter.StartWS(context.Background()); err != nil {
				logger.Error("Failed to start Feishu WS", zap.Error(err))
			}
		}()
	} else {
		logger.Warn("Feishu AppID or AppSecret is empty, skipping Feishu adapter start")
	}

	// 7.2 REST API Adapter (For Coze, Dify, Custom Webhooks)
	// This also serves as the HTTP server
	port := config.AppConfig.Server.Port
	restAdapter := rest.NewAdapter(port, dispatcher, logger)

	// TODO: 7.3 Add WeChat Adapter
	// wechatAdapter := wechat.NewAdapter(..., dispatcher, logger)
	// go wechatAdapter.Start()

	// TODO: 7.4 Add Telegram Adapter
	// telegramAdapter := telegram.NewAdapter(..., dispatcher, logger)
	// go telegramAdapter.Start()

	// Start REST Server (Blocking, or handle gracefully)
	go func() {
		if err := restAdapter.Start(context.Background()); err != nil {
			log.Fatalf("REST Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")
}
