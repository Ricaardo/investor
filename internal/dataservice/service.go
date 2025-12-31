package dataservice

import (
	"context"
	"math"
	"math/rand"
	"time"
)

type DataService interface {
	GetIPOList(ctx context.Context) ([]IPOInfo, error)
	GetMarketQuote(ctx context.Context, symbol string) (*MarketQuote, error)
	SearchMarketNews(ctx context.Context, query string) ([]NewsItem, error)
	GetMarketIndex(ctx context.Context) ([]IndexQuote, error)
	GetSecurityAnalysis(ctx context.Context, symbol string, assetType string) (*SecurityAnalysis, error)
	GetHistoricalQuotes(ctx context.Context, symbol string, interval string, rangeStr string) ([]KLineItem, error)
	GetMarketSentiment(ctx context.Context, market string) (*SentimentData, error)
}

type IPOInfo struct {
	Name             string  `json:"name"`
	Code             string  `json:"code"`
	Price            float64 `json:"price"`
	ListingDate      string  `json:"listing_date"`
	SubscriptionDate string  `json:"subscription_date"`
}

type MarketQuote struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"change_pct"`
	UpdatedAt string  `json:"updated_at"`
}

type NewsItem struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Source  string `json:"source"`
	Time    string `json:"time"`
}

type IndexQuote struct {
	Name      string  `json:"name"`
	Value     float64 `json:"value"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"change_pct"`
}

type SentimentData struct {
	Market      string  `json:"market"`      // "crypto", "us_stock", "cn_stock"
	Score       float64 `json:"score"`       // 0-100 (0=Extreme Fear, 100=Extreme Greed)
	Label       string  `json:"label"`       // "Fear", "Greed", "Neutral", etc.
	Description string  `json:"description"` // Description or reason
	Timestamp   int64   `json:"timestamp"`
}

// SecurityAnalysis contains deep analysis data for LLM
type SecurityAnalysis struct {
	Symbol          string      `json:"symbol"`
	AssetType       string      `json:"asset_type"` // stock, crypto, forex, commodity
	CurrentPrice    float64     `json:"current_price"`
	MA20            float64     `json:"ma20"`
	MA60            float64     `json:"ma60"`
	RSI             float64     `json:"rsi"`       // 14-day RSI
	VolumeRatio     float64     `json:"vol_ratio"` // Today vol / Avg vol
	Trend           string      `json:"trend"`     // "bullish", "bearish", "sideways"
	SupportLevel    float64     `json:"support"`
	ResistanceLevel float64     `json:"resistance"`
	RecentKLines    []KLineItem `json:"recent_klines"` // Last 5 days for context
}

type KLineItem struct {
	Date   string  `json:"date"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

// MockDataService implements DataService with mock data
type MockDataService struct{}

func NewMockDataService() *MockDataService {
	return &MockDataService{}
}

func (s *MockDataService) GetIPOList(ctx context.Context) ([]IPOInfo, error) {
	return []IPOInfo{
		{Name: "阿里云", Code: "BABA-SW", Price: 88.8, ListingDate: "2025-02-01", SubscriptionDate: "2025-01-25"},
		{Name: "字节跳动", Code: "BYTE", Price: 120.5, ListingDate: "2025-03-15", SubscriptionDate: "2025-03-01"},
		{Name: "蜜雪冰城", Code: "MXBC", Price: 15.2, ListingDate: "2025-01-10", SubscriptionDate: "2025-01-05"},
	}, nil
}

func (s *MockDataService) GetMarketQuote(ctx context.Context, symbol string) (*MarketQuote, error) {
	rand.Seed(time.Now().UnixNano())
	basePrice := 100.0 + rand.Float64()*500
	change := (rand.Float64() - 0.5) * 10

	return &MarketQuote{
		Symbol:    symbol,
		Price:     basePrice,
		Change:    change,
		ChangePct: change / basePrice * 100,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *MockDataService) SearchMarketNews(ctx context.Context, query string) ([]NewsItem, error) {
	// Mock news data
	news := []NewsItem{
		{Title: "美联储暗示明年可能降息", Summary: "在最新的FOMC会议纪要中，美联储官员讨论了通胀下降的趋势。", Source: "财联社", Time: "10:30"},
		{Title: "比特币突破10万美元大关", Summary: "受ETF资金持续流入影响，加密货币市场全线大涨。", Source: "Coindesk", Time: "08:15"},
		{Title: "黄金价格创历史新高", Summary: "避险情绪升温，现货黄金突破2200美元/盎司。", Source: "Bloomberg", Time: "14:20"},
	}

	return news, nil
}

func (s *MockDataService) GetMarketIndex(ctx context.Context) ([]IndexQuote, error) {
	return []IndexQuote{
		{Name: "上证指数", Value: 3050.23, Change: 15.5, ChangePct: 0.51},
		{Name: "纳斯达克", Value: 14500.12, Change: -50.2, ChangePct: -0.35},
		{Name: "恒生指数", Value: 17000.50, Change: 120.8, ChangePct: 0.71},
		{Name: "BTC/USD", Value: 42000.00, Change: 1200.0, ChangePct: 2.9},
		{Name: "Gold (XAU)", Value: 2050.50, Change: 10.5, ChangePct: 0.5},
	}, nil
}

func (s *MockDataService) GetSecurityAnalysis(ctx context.Context, symbol string, assetType string) (*SecurityAnalysis, error) {
	rand.Seed(time.Now().UnixNano())

	// Mock base price based on asset type
	basePrice := 100.0
	switch assetType {
	case "crypto":
		basePrice = 40000.0 // Like BTC
	case "gold":
		basePrice = 2000.0
	case "forex":
		basePrice = 1.08 // Like EUR/USD
	default:
		basePrice = 150.0 // Stock
	}

	// Add randomness
	currentPrice := basePrice * (1 + (rand.Float64()-0.5)*0.1)

	// Mock Technical Indicators
	ma20 := currentPrice * (1 + (rand.Float64()-0.5)*0.05)
	ma60 := currentPrice * (1 + (rand.Float64()-0.5)*0.1)
	rsi := 30.0 + rand.Float64()*40.0 // 30-70 range mostly

	trend := "sideways"
	if ma20 > ma60 && currentPrice > ma20 {
		trend = "bullish"
	} else if ma20 < ma60 && currentPrice < ma20 {
		trend = "bearish"
	}

	// Mock KLines
	klines := make([]KLineItem, 5)
	for i := 0; i < 5; i++ {
		klines[i] = KLineItem{
			Date:   time.Now().AddDate(0, 0, -5+i).Format("2006-01-02"),
			Close:  currentPrice * (1 + (rand.Float64()-0.5)*0.02),
			Volume: 1000000 * rand.Float64(),
		}
	}

	return &SecurityAnalysis{
		Symbol:          symbol,
		AssetType:       assetType,
		CurrentPrice:    math.Round(currentPrice*100) / 100,
		MA20:            math.Round(ma20*100) / 100,
		MA60:            math.Round(ma60*100) / 100,
		RSI:             math.Round(rsi*100) / 100,
		VolumeRatio:     0.8 + rand.Float64()*0.4,
		Trend:           trend,
		SupportLevel:    math.Round(currentPrice*0.9*100) / 100,
		ResistanceLevel: math.Round(currentPrice*1.1*100) / 100,
		RecentKLines:    klines,
	}, nil
}

func (s *MockDataService) GetHistoricalQuotes(ctx context.Context, symbol string, interval string, rangeStr string) ([]KLineItem, error) {
	// Mock KLines
	klines := make([]KLineItem, 10)
	basePrice := 100.0
	for i := 0; i < 10; i++ {
		klines[i] = KLineItem{
			Date:   time.Now().AddDate(0, 0, -10+i).Format("2006-01-02"),
			Close:  basePrice * (1 + (rand.Float64()-0.5)*0.1),
			Volume: 1000000 * rand.Float64(),
		}
	}
	return klines, nil
}

func (s *MockDataService) GetMarketSentiment(ctx context.Context, market string) (*SentimentData, error) {
	return &SentimentData{
		Market:      market,
		Score:       65,
		Label:       "Greed",
		Description: "Market is showing signs of greed due to recent rally.",
		Timestamp:   time.Now().Unix(),
	}, nil
}

// Function definitions for LLM Tool Calling
var ToolsDefinition = []map[string]interface{}{
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "get_ipo_list",
			"description": "获取近期即将上市的新股(IPO)列表信息",
			"parameters": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "get_market_quote",
			"description": "获取指定标的的实时行情价格 (支持股票、加密货币、外汇、贵金属)",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"symbol": map[string]interface{}{
						"type":        "string",
						"description": "代码，如 'AAPL', 'BTC-USD', 'XAU', 'EURUSD'",
					},
				},
				"required": []string{"symbol"},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "search_market_news",
			"description": "搜索最新的财经新闻资讯 (支持关键词或特定类别)",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "搜索关键词(如 'Tesla', 'CPI', '降息') 或 类别('macro', 'crypto', 'us_stock', 'cn_stock')",
					},
				},
				"required": []string{"query"},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "get_market_index",
			"description": "获取主要市场指数（如上证、纳指、BTC、黄金）",
			"parameters": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "get_security_analysis",
			"description": "获取标的的深度技术分析数据（含均线、RSI、趋势判断、支撑压力位、量能分析）",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"symbol": map[string]interface{}{
						"type":        "string",
						"description": "标的代码，如 'BTC', 'ETH', 'Gold', 'AAPL'",
					},
					"asset_type": map[string]interface{}{
						"type":        "string",
						"description": "资产类别: 'stock'(股票), 'crypto'(加密货币), 'gold'(黄金/贵金属), 'forex'(外汇), 'commodity'(商品)",
						"enum":        []string{"stock", "crypto", "gold", "forex", "commodity"},
					},
				},
				"required": []string{"symbol", "asset_type"},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "get_market_sentiment",
			"description": "获取市场恐慌与贪婪指数 (Crypto/Stock)",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"market": map[string]interface{}{
						"type":        "string",
						"description": "市场类型: 'crypto'(加密货币), 'us_stock'(美股)",
						"enum":        []string{"crypto", "us_stock"},
					},
				},
				"required": []string{"market"},
			},
		},
	},
}
