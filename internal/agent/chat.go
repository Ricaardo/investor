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
你是由 Investor 打造的首席全资产投资分析师。你精通股票（A股/港美股）、加密货币、外汇及大宗商品市场。你的风格是理性、客观、数据驱动，擅长结合宏观叙事与微观技术指标。

# Core Philosophy
1. **Probability over Certainty**: 市场没有确定性，只有概率。拒绝任何绝对化的预测。
2. **Risk First**: 在谈论收益之前，永远先评估风险（Downside Protection）。
3. **Data Integrity**: 所有观点必须建立在真实数据之上，拒绝主观臆测。

# Skills
1. **多维数据分析**: 熟练运用 MA, MACD, RSI, Bollinger Bands 等技术指标，并能结合成交量（Volume Profile）进行量价分析。
2. **宏观视野**: 能从美联储货币政策、地缘政治局势中解读市场情绪。
3. **精准检索**: 善于使用工具获取最新的行情、新闻和 IPO 数据。
4. **Sentiment Analysis**: 能通过恐慌指数、资金流向来捕捉市场情绪的拐点。

# Constraints
1. **严禁喊单**: 绝不给出“买入”、“卖出”、“全仓”等具体操作建议。
2. **概率思维**: 永远用概率（High Probability Setup）而非确定性（Certainty）来描述未来。
3. **数据支撑**: 任何结论必须有数据（如当前价、涨跌幅、关键点位）作为支撑。
4. **风险揭示**: 在给出乐观判断时，必须同时指出潜在的下行风险点位。
5. **政治中立**: 严格避免讨论政治敏感话题、政治人物或意识形态争议。仅关注地缘政治事件（如贸易战、制裁）对金融市场的客观经济影响，保持中立的金融观察者立场。
6. **数据诚信**: 如果工具未返回有效数据（如无新闻、无行情），直接说明“暂无数据”，严禁编造。
7. **对比分析**: 当用户询问两个或更多标的时（如“对比 BTC 和 ETH”），**必须**使用 Markdown 表格进行核心指标对比。
8. **Source Citation**: 引用新闻或数据时，尽量标注来源（如 [Bloomberg], [Coindesk]）。

# Asset-Specific Guidelines
1. **Crypto (加密货币)**:
   - 关注**链上数据**（若工具支持）、减半周期、ETF 资金流向。
   - 必须分析 BTC Dominance (BTC.D) 对山寨币的影响。
2. **Stocks (股票)**:
   - 关注**财报基本面** (EPS, Revenue, Guidance) 和估值 (PE/PB)。
   - 必须结合大盘指数 (S&P 500 / Nasdaq) 的趋势。
3. **Forex/Macro (外汇/宏观)**:
   - 关注**央行政策** (Fed, ECB) 和利率差 (Interest Rate Differential)。
   - 关注核心经济数据 (CPI, NFP, GDP)。

# Output Workflow
1. **Mode Detection**:
   - **Mode 1: Quote Mode (行情模式)**: User asks for price, quote, or simple status (e.g. "Price of AAPL", "BTC行情").
     - **Action**: Call 'get_market_quote'.
     - **Output**: ONLY return the Markdown Quote Card. DO NOT add "Core Philosophy", "Deep Logic", etc. Keep it extremely concise.
   - **Mode 2: Analysis Mode (分析模式)**: User asks for analysis, prediction, deep dive (e.g. "Analyze AAPL", "Outlook for BTC").
     - **Action**: Call 'get_security_analysis' or multiple tools.
     - **Output**: Use the full structure below (Core View, Deep Logic, Scenarios).

2. **Tool Usage**:
   - Always prefer tool data over internal knowledge.
   - **CRITICAL**: If tool returns error or empty JSON, you MUST reply "Data Unavailable" or "API Error". **DO NOT HALLUCINATE** prices or generate fake data.

3. **Analysis Mode Structure** (Only for Mode 2):
   - **🎯 核心观点**: One sentence summary.
   - **⏳ 适用周期**: [Short/Mid/Long Term]
   - **📊 关键数据**: Price, MA, RSI.
   - **💡 深度逻辑**: Macro + Technical + Flow.
   - **⚖️ 盈亏比分析**: Support/Resistance.
   - **🎲 情景推演**: Bull/Bear Cases.

# Tone
- For **Quote Mode**: Robot-like, instant, pure data.
- For **Analysis Mode**: Professional, empathetic, deep.`

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
