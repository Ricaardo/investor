package feishu

import (
	"context"
	"encoding/json"
	
	"investot/config"
	"investot/internal/core"
	"investot/internal/model"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
	"go.uber.org/zap"
)

type Adapter struct {
	Config     config.FeishuConfig
	Dispatcher *core.Dispatcher
	Logger     *zap.Logger
	Client     *lark.Client
}

func NewAdapter(cfg config.FeishuConfig, dispatcher *core.Dispatcher, logger *zap.Logger) *Adapter {
	client := lark.NewClient(cfg.AppID, cfg.AppSecret,
		lark.WithLogReqAtDebug(true),
		lark.WithLogLevel(larkcore.LogLevelDebug),
	)

	return &Adapter{
		Config:     cfg,
		Dispatcher: dispatcher,
		Logger:     logger,
		Client:     client,
	}
}

// StartWS starts the WebSocket connection
func (a *Adapter) StartWS(ctx context.Context) error {
	// Use larkevent.NewEventDispatcher for WS event handling
	// Note: For WS, we use "larkevent" package alias which now points to "github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	eventHandler := larkevent.NewEventDispatcher(a.Config.VerificationToken, a.Config.EncryptKey).
		OnP2MessageReceiveV1(a.handleMessage).
		OnP2MessageReadV1(func(ctx context.Context, event *larkim.P2MessageReadV1) error {
			// Handle read receipt if needed
			return nil
		})

	cli := larkws.NewClient(a.Config.AppID, a.Config.AppSecret,
		larkws.WithEventHandler(eventHandler),
		larkws.WithLogLevel(larkcore.LogLevelDebug),
	)

	a.Logger.Info("Starting Feishu WebSocket client...")
	return cli.Start(ctx)
}

func (a *Adapter) handleMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	content := event.Event.Message.Content
	msgID := event.Event.Message.MessageId
	chatID := event.Event.Message.ChatId
	senderID := event.Event.Sender.SenderId.OpenId
	
	// Parse JSON content: {"text":"hello"}
	var contentMap map[string]string
	if err := json.Unmarshal([]byte(*content), &contentMap); err != nil {
		a.Logger.Error("Failed to parse message content", zap.Error(err))
		return nil
	}
	text := contentMap["text"]

	a.Logger.Info("Received message", zap.String("text", text), zap.String("sender", *senderID))

	// 1. Convert to Internal Message
	internalMsg := model.InternalMessage{
		Platform:    "feishu",
		ChatType:    "private", // TODO: Detect group
		ChatID:      *chatID,
		UserID:      *senderID,
		Text:        text,
		Timestamp:   0, // Lark event doesn't have simple timestamp in root, can ignore for now
		IsMentioned: false,
	}

	// 2. Dispatch
	go func() {
		response, err := a.Dispatcher.Dispatch(context.Background(), &internalMsg)
		if err != nil {
			a.Logger.Error("Dispatch failed", zap.Error(err))
			return
		}

		if response != "" {
			a.Reply(*msgID, response)
		}
	}()

	return nil
}

func (a *Adapter) Reply(messageID string, text string) {
	// Properly marshal content to JSON to handle escaping (newlines, quotes, etc.)
	contentMap := map[string]string{
		"text": text,
	}
	contentBytes, err := json.Marshal(contentMap)
	if err != nil {
		a.Logger.Error("Failed to marshal reply content", zap.Error(err))
		return
	}
	content := string(contentBytes)

	resp, err := a.Client.Im.Message.Reply(context.Background(), larkim.NewReplyMessageReqBuilder().
		MessageId(messageID).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeText).
			Content(content).
			Build()).
		Build())

	if err != nil {
		a.Logger.Error("Failed to reply message", zap.Error(err))
		return
	}

	if !resp.Success() {
		a.Logger.Error("Failed to reply message (API error)", zap.Int("code", resp.Code), zap.String("msg", resp.Msg))
	} else {
		a.Logger.Info("Reply sent to Feishu")
	}
}
