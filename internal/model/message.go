package model

type InternalMessage struct {
	Platform    string `json:"platform"`     // "wecom", "feishu"
	ChatType    string `json:"chat_type"`    // "private", "group"
	ChatID      string `json:"chat_id"`      // Conversation ID
	UserID      string `json:"user_id"`      // Sender ID
	Username    string `json:"username"`     // Sender Name
	Text        string `json:"text"`         // Message Content
	IsMentioned bool   `json:"is_mentioned"` // Is mentioned?
	Timestamp   int64  `json:"timestamp"`
}

type ReplyMessage struct {
	Platform string `json:"platform"`
	ChatID   string `json:"chat_id"`
	Text     string `json:"text"`
}
