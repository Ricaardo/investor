package dataservice

import (
	"fmt"
	"strings"
	"time"
)

// ToMarkdown formats MarketQuote to markdown
func (m *MarketQuote) ToMarkdown() string {
	icon := "📈"
	if m.Change < 0 {
		icon = "📉"
	}
	// Handle nil time or parse error
	tStr := m.UpdatedAt
	if t, err := time.Parse(time.RFC3339, m.UpdatedAt); err == nil {
		tStr = t.Format("2006-01-02 15:04:05")
	}

	// Determine Chart Link
	chartLink := fmt.Sprintf("https://finance.yahoo.com/quote/%s/chart", m.Symbol)

	// Special handling for Crypto
	if strings.Contains(strings.ToUpper(m.Symbol), "USDT") {
		// If symbol has hyphen (e.g. BTC-USDT from OKX), link to OKX
		if strings.Contains(m.Symbol, "-") {
			chartLink = fmt.Sprintf("https://www.okx.com/zh-hans/trade-spot/%s", strings.ToLower(m.Symbol))
		} else {
			// Default to TradingView (Binance data)
			chartLink = fmt.Sprintf("https://www.tradingview.com/chart/?symbol=BINANCE:%s", m.Symbol)
		}
	}

	return fmt.Sprintf("📊 **%s 实时行情**\n-------------------\n💰 价格: %.2f\n%s 涨跌: %.2f (%.2f%%)\n⏰ 更新: %s\n🔗 [查看K线图表](%s)",
		m.Symbol, m.Price, icon, m.Change, m.ChangePct, tStr, chartLink)
}

// GenerateSparkline creates a unicode sparkline from data
func GenerateSparkline(data []float64) string {
	if len(data) == 0 {
		return ""
	}
	min := data[0]
	max := data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	rangeVal := max - min
	if rangeVal == 0 {
		return strings.Repeat("▅", len(data))
	}

	blocks := []string{" ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	var sb strings.Builder
	for _, v := range data {
		idx := int((v - min) / rangeVal * float64(len(blocks)-1))
		sb.WriteString(blocks[idx])
	}
	return sb.String()
}

// ToMarkdown formats SecurityAnalysis to markdown
func (s *SecurityAnalysis) ToMarkdown() string {
	trendIcon := "➡️"
	if s.Trend == "bullish" {
		trendIcon = "🐂"
	} else if s.Trend == "bearish" {
		trendIcon = "🐻"
	}

	// Generate Sparkline from RecentKLines
	var closes []float64
	for _, k := range s.RecentKLines {
		closes = append(closes, k.Close)
	}
	sparkline := GenerateSparkline(closes)
	if sparkline != "" {
		sparkline = "\n📈 走势: " + sparkline
	}

	return fmt.Sprintf(`🔍 **%s 深度技术分析**
-------------------
当前价: %.2f | 趋势: %s %s%s
-------------------
• **均线系统**:
  MA20: %.2f
  MA60: %.2f
• **技术指标**:
  RSI(14): %.2f
  量比: %.2f
• **关键点位**:
  压力位: %.2f
  支撑位: %.2f
-------------------
*注: 以上数据仅供参考，不构成投资建议*`,
		s.Symbol, s.CurrentPrice, trendIcon, s.Trend, sparkline,
		s.MA20, s.MA60,
		s.RSI, s.VolumeRatio,
		s.ResistanceLevel, s.SupportLevel)
}

// ToMarkdownList formats a slice of NewsItem to markdown
func ToMarkdownNewsList(news []NewsItem) string {
	if len(news) == 0 {
		return "暂无相关新闻资讯。"
	}

	var sb strings.Builder
	sb.WriteString("📰 **最新市场资讯**\n-------------------\n")

	for i, n := range news {
		if i >= 5 {
			break
		}

		sb.WriteString(fmt.Sprintf("• **%s**\n", n.Title))
		sb.WriteString(fmt.Sprintf("  *来源: %s | 时间: %s*\n", n.Source, n.Time))
		if n.Summary != "" {
			// Truncate summary
			summary := n.Summary
			if len(summary) > 100 {
				summary = summary[:100] + "..."
			}
			sb.WriteString(fmt.Sprintf("  > %s\n", summary))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
