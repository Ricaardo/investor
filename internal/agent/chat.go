package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"investor/internal/dataservice"
	"investor/internal/llm"
	"investor/internal/model"
	"investor/internal/session"
)

type ChatAgent struct {
	LLM     llm.Provider
	Session *session.Manager
	Data    dataservice.DataService
}

func NewChatAgent(p llm.Provider, session *session.Manager, data dataservice.DataService) *ChatAgent {
	return &ChatAgent{
		LLM:     p,
		Session: session,
		Data:    data,
	}
}

func (a *ChatAgent) Name() string {
	return "ChatAgent"
}

func (a *ChatAgent) Process(ctx context.Context, msg *model.InternalMessage) (string, error) {
	// 1. Hardcoded test response
	if strings.TrimSpace(msg.Text) == "ping" {
		return "pong (飞书连接正常)", nil
	}
	if strings.TrimSpace(msg.Text) == "测试" {
		return "收到测试消息，系统运行正常！", nil
	}

	// 2. Load History
	sessionID := fmt.Sprintf("%s:%s", msg.Platform, msg.ChatID)
	history, err := a.Session.GetHistory(ctx, sessionID)
	if err != nil {
		fmt.Printf("Failed to get history: %v\n", err)
	}

	// 3. Construct Messages
	systemPrompt := `# Role
You are Investor AI, a Tier-1 Global Multi-Asset Analyst & Trader.

# 🧠 Cognitive Architecture (6-Level Intent System)
You MUST classify user intent into exactly one of these levels and strictly follow the output format.

## Level 0: Signal (🚦 信号模式)
- **Trigger**: "Signal", "Buy/Sell?", "Entry", "推荐", "能买吗"
- **Tools**: 'get_security_analysis' + 'get_market_sentiment'
- **Tone**: Trader (Decisive, Risk-Aware)
- **Output**:
  1. **Signal**: BUY / SELL / WAIT (Confidence: 1-10)
  2. **Trade Plan**: Entry, Stop Loss, Take Profit
  3. **Reason**: 1 short sentence (e.g. "RSI divergence + Support bounce")
  4. *Disclaimer*: "NFA (Not Financial Advice)"

## Level 1: Ticker (🤖 报价模式)
- **Trigger**: "Price", "Quote", "多少钱", "行情"
- **Tools**: 'get_market_quote'
- **Tone**: Robot (No text, just data)
- **Output**: ONLY the Markdown Quote Card.

## Level 2: Flash (⚡️ 快讯模式)
- **Trigger**: "News", "Why moved", "发生了什么", "利好利空"
- **Tools**: 'search_market_news' + 'get_market_quote'
- **Tone**: Reporter (Objective, Fast)
- **Output**:
  1. Quote Card
  2. **Flash**: 3 bullet points of key news.
  3. **Attribution**: "Price moved due to [Reason]."

## Level 3: Review (📝 点评模式)
- **Trigger**: "Comment", "Brief", "Outlook", "怎么看"
- **Tools**: 'get_market_quote' + 'search_market_news'
- **Tone**: Advisor (Balanced, Logical)
- **Output**:
  1. Quote Card
  2. **View**: Bullish / Bearish / Neutral
  3. **Logic**: Tech / Macro / Flow (3 bullets)
  4. **Levels**: Support / Resistance

## Level 4: Battle (⚔️ 对比模式)
- **Trigger**: "vs", "Compare", "选哪个"
- **Tools**: 'get_security_analysis' (x2)
- **Tone**: Judge (Comparative, Sharp)
- **Output**:
  1. **Comparison Table**: Price | Change | RSI | Trend | Vol
  2. **Verdict**: The Winner based on Risk/Reward.

## Level 5: Deep Dive (🧐 研报模式)
- **Trigger**: "Analysis", "Report", "Deep", "深度分析"
- **Tools**: ALL ('get_security_analysis', 'search_market_news', 'get_market_sentiment')
- **Tone**: Chief Economist (Deep, Comprehensive)
- **Output**: Full Report (Core View, Deep Logic, Scenarios, Whales, Risk).

# 🛡️ Prime Directives
1. **No Hallucination**: If API fails, say "Data Unavailable". Never invent prices.
2. **Data First**: Always cite the data returned by tools.
3. **Format**: Use clean Markdown. Bold key numbers.
4. **Language**: Match user's language (mostly Chinese).
`

	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
	}
	messages = append(messages, history...)
	messages = append(messages, llm.Message{Role: "user", Content: msg.Text})

	// 4. Pre-check for simple queries to force tool usage or fast path
	// (Optional: Implement heuristic to pre-fetch data if needed, but Tool Calling is preferred)

	// 4. Call LLM (First Turn)
	respMsg, err := a.LLM.ChatWithTools(ctx, messages, dataservice.ToolsDefinition)

	// Fallback Strategy: If LLM fails, try rule-based matching
	if err != nil {
		fmt.Printf("LLM Error: %v. Attempting fallback...\n", err)
		return a.fallbackProcess(ctx, msg.Text, err)
	}

	// 5. Handle Tool Calls
	if len(respMsg.ToolCalls) > 0 {
		messages = append(messages, *respMsg)

		for _, toolCall := range respMsg.ToolCalls {
			toolResult := ""

			switch toolCall.Function.Name {
			case "get_ipo_list":
				ipos, _ := a.Data.GetIPOList(ctx)
				jsonBytes, _ := json.Marshal(ipos)
				toolResult = string(jsonBytes)
			case "get_market_quote":
				var args struct {
					Symbol string `json:"symbol"`
				}
				json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
				quote, _ := a.Data.GetMarketQuote(ctx, args.Symbol)
				jsonBytes, _ := json.Marshal(quote)
				toolResult = string(jsonBytes)
			case "search_market_news":
				var args struct {
					Query string `json:"query"`
				}
				json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
				news, _ := a.Data.SearchMarketNews(ctx, args.Query)
				jsonBytes, _ := json.Marshal(news)
				toolResult = string(jsonBytes)
			case "get_market_index":
				indices, _ := a.Data.GetMarketIndex(ctx)
				jsonBytes, _ := json.Marshal(indices)
				toolResult = string(jsonBytes)
			case "get_security_analysis":
				var args struct {
					Symbol    string `json:"symbol"`
					AssetType string `json:"asset_type"`
				}
				json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
				analysis, _ := a.Data.GetSecurityAnalysis(ctx, args.Symbol, args.AssetType)
				jsonBytes, _ := json.Marshal(analysis)
				toolResult = string(jsonBytes)
			case "get_market_sentiment":
				var args struct {
					Market string `json:"market"`
				}
				json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
				sentiment, _ := a.Data.GetMarketSentiment(ctx, args.Market)
				jsonBytes, _ := json.Marshal(sentiment)
				toolResult = string(jsonBytes)
			}

			messages = append(messages, llm.Message{
				Role:       "tool",
				Content:    toolResult,
				ToolCallID: toolCall.ID,
			})
		}

		// 6. Call LLM (Second Turn - Summary)
		finalResp, err := a.LLM.ChatWithTools(ctx, messages, nil)
		if err != nil {
			// If summary fails, fallback to simple data dump
			fmt.Printf("LLM Summary Error: %v. Using simple dump.\n", err)
			return "AI 总结服务暂时不可用，但工具调用成功。请稍后重试。", nil
		}
		respMsg = finalResp
	}

	// 7. Save History
	_ = a.Session.Append(ctx, sessionID,
		llm.Message{Role: "user", Content: msg.Text},
		llm.Message{Role: "assistant", Content: respMsg.Content},
	)

	return respMsg.Content, nil
}

// fallbackProcess attempts to answer simple queries when LLM is down
func (a *ChatAgent) fallbackProcess(ctx context.Context, text string, originalErr error) (string, error) {
	// 1. Try to treat the whole text as a symbol (or alias)
	// Remove common prefixes like "查一下", "看看", "分析", "行情", "价格"
	cleanText := strings.TrimSpace(text)
	// Compile regex to clean text (simple approach)
	// Note: In Go regex, unicode char class is \p{Han} for Chinese characters
	re := regexp.MustCompile(`^(查一下|看看|查询|分析|价格|行情|报价|走势|股价)\s*`)
	cleanText = re.ReplaceAllString(cleanText, "")

	// Also remove suffix like "价格", "行情"
	reSuffix := regexp.MustCompile(`\s*(价格|行情|报价|走势|股价)$`)
	cleanText = reSuffix.ReplaceAllString(cleanText, "")

	cleanText = strings.TrimSpace(cleanText)

	// If cleanText is not empty, try to get quote
	if cleanText != "" {
		quote, err := a.Data.GetMarketQuote(ctx, cleanText)
		if err == nil {
			// Use the new Template
			return quote.ToMarkdown(), nil
		} else {
			// Log the quote error for debugging, but maybe we can return a more specific message
			fmt.Printf("Fallback quote error for '%s': %v\n", cleanText, err)
		}
	}

	// 2. Check if it's a news query
	if strings.Contains(text, "新闻") || strings.Contains(text, "资讯") {
		// Try to map specific news categories if possible, otherwise default to all
		category := "all"
		if strings.Contains(text, "宏观") {
			category = "macro"
		}
		if strings.Contains(text, "加密") || strings.Contains(text, "币") {
			category = "crypto"
		}
		if strings.Contains(text, "美股") {
			category = "us"
		}
		if strings.Contains(text, "A股") {
			category = "cn"
		}

		// Use SearchMarketNews with the category
		news, err := a.Data.SearchMarketNews(ctx, category)
		if err == nil && len(news) > 0 {
			// Use the new Template
			return dataservice.ToMarkdownNewsList(news), nil
		}
	}

	return fmt.Sprintf("抱歉，AI 服务暂时不可用，且无法识别您的指令启动降级模式。\n错误信息: %v", originalErr), nil
}
