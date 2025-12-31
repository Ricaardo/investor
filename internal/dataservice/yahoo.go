package dataservice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/piquette/finance-go/quote"
)

type YahooDataService struct{}

func NewYahooDataService() *YahooDataService {
	return &YahooDataService{}
}

func (s *YahooDataService) GetIPOList(ctx context.Context) ([]IPOInfo, error) {
	return []IPOInfo{
		{Name: "Mock: NextEra", Code: "NEXT", Price: 25.0, ListingDate: "2025-02-15", SubscriptionDate: "2025-02-10"},
	}, nil
}

// YahooSearchResponse for Autocomplete API
type YahooSearchResponse struct {
	Count  int `json:"count"`
	Quotes []struct {
		Symbol    string `json:"symbol"`
		Shortname string `json:"shortname"`
		Longname  string `json:"longname"`
		Exchange  string `json:"exchange"`
		TypeDisp  string `json:"typeDisp"`
	} `json:"quotes"`
}

// searchYahooSymbol uses Yahoo Autocomplete API to find symbol by name
func searchYahooSymbol(query string) string {
	apiURL := fmt.Sprintf("https://query2.finance.yahoo.com/v1/finance/search?q=%s&lang=zh-CN&region=CN&quotesCount=1&newsCount=0", url.QueryEscape(query))

	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequest("GET", apiURL, nil)
	// User-Agent to avoid 403
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ""
	}

	body, _ := io.ReadAll(resp.Body)
	var result YahooSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return ""
	}

	if len(result.Quotes) > 0 {
		return result.Quotes[0].Symbol
	}
	return ""
}

// Yahoo Chart V8 Response
type YahooChartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				RegularMarketPrice float64 `json:"regularMarketPrice"`
				PreviousClose      float64 `json:"previousClose"`
				ChartPreviousClose float64 `json:"chartPreviousClose"` // Added this field
				Symbol             string  `json:"symbol"`
				RegularMarketTime  int64   `json:"regularMarketTime"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []int64   `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}

func getYahooPriceV8(symbol string) (*MarketQuote, error) {
	// Use Yahoo Chart API V8 as it is more stable than Quote API
	apiURL := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1d&range=1d", symbol)

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", apiURL, nil)
	// Critical: User-Agent to avoid 403/429
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("yahoo chart api error: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var chartResp YahooChartResponse
	if err := json.Unmarshal(body, &chartResp); err != nil {
		return nil, err
	}

	if len(chartResp.Chart.Result) == 0 {
		return nil, fmt.Errorf("symbol not found or no data: %s", symbol)
	}

	meta := chartResp.Chart.Result[0].Meta
	price := meta.RegularMarketPrice

	// Validate price data
	if price == 0 && meta.PreviousClose == 0 {
		return nil, fmt.Errorf("invalid price data (0.0) for symbol: %s", symbol)
	}

	// Fallback to ChartPreviousClose if PreviousClose is missing/zero
	prevClose := meta.PreviousClose
	if prevClose == 0 {
		prevClose = meta.ChartPreviousClose
	}

	change := price - prevClose
	changePct := 0.0
	if prevClose != 0 {
		changePct = (change / prevClose) * 100
	}

	return &MarketQuote{
		Symbol:    meta.Symbol,
		Price:     price,
		Change:    change,
		ChangePct: changePct,
		UpdatedAt: time.Unix(meta.RegularMarketTime, 0).Format(time.RFC3339),
	}, nil
}

// Binance API response
type BinanceTicker struct {
	Symbol             string `json:"symbol"`
	LastPrice          string `json:"lastPrice"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
}

func getBinancePrice(symbol string) (*MarketQuote, error) {
	resp, err := http.Get("https://api.binance.com/api/v3/ticker/24hr?symbol=" + strings.ToUpper(symbol))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("binance api error: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var ticker BinanceTicker
	if err := json.Unmarshal(body, &ticker); err != nil {
		return nil, err
	}

	price, _ := strconv.ParseFloat(ticker.LastPrice, 64)
	change, _ := strconv.ParseFloat(ticker.PriceChange, 64)
	changePct, _ := strconv.ParseFloat(ticker.PriceChangePercent, 64)

	return &MarketQuote{
		Symbol:    ticker.Symbol,
		Price:     price,
		Change:    change,
		ChangePct: changePct,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// OKX API response
type OkxTicker struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		InstID  string `json:"instId"`
		Last    string `json:"last"`
		Open24h string `json:"open24h"`
	} `json:"data"`
}

func getOkxPrice(symbol string) (*MarketQuote, error) {
	// Convert BTCUSDT -> BTC-USDT
	if !strings.Contains(symbol, "-") && strings.HasSuffix(symbol, "USDT") {
		symbol = strings.TrimSuffix(symbol, "USDT") + "-USDT"
	}
	symbol = strings.ToUpper(symbol)

	resp, err := http.Get("https://www.okx.com/api/v5/market/ticker?instId=" + symbol)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("okx api error: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var ticker OkxTicker
	if err := json.Unmarshal(body, &ticker); err != nil {
		return nil, err
	}

	if ticker.Code != "0" || len(ticker.Data) == 0 {
		return nil, fmt.Errorf("okx api error: %s", ticker.Msg)
	}

	data := ticker.Data[0]
	price, _ := strconv.ParseFloat(data.Last, 64)
	open, _ := strconv.ParseFloat(data.Open24h, 64)

	change := price - open
	changePct := 0.0
	if open != 0 {
		changePct = (change / open) * 100
	}

	return &MarketQuote{
		Symbol:    data.InstID,
		Price:     price,
		Change:    change,
		ChangePct: changePct,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *YahooDataService) GetMarketQuote(ctx context.Context, symbol string) (*MarketQuote, error) {
	symbol = normalizeSymbol(symbol)

	// Strategy 1: Binance for Crypto
	if strings.HasSuffix(strings.ToUpper(symbol), "USDT") {
		// Try OKX First (User requested OKX fix/support)
		q, err := getOkxPrice(symbol)
		if err == nil {
			return q, nil
		}
		// Fallback to Binance
		q, err = getBinancePrice(symbol)
		if err == nil {
			return q, nil
		}
	}

	// Strategy 2: Yahoo Finance Chart API V8 (Custom Implementation)
	// We prefer V8 Chart API over finance-go/quote because V8 is more robust against 403/429
	q, err := getYahooPriceV8(symbol)
	if err == nil {
		return q, nil
	}

	// Strategy 3: Fallback to finance-go (Old method, likely to fail if V8 failed)
	oldQ, err := quote.Get(symbol)
	if err != nil {
		return nil, err
	}
	if oldQ == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbol)
	}

	return &MarketQuote{
		Symbol:    symbol,
		Price:     oldQ.RegularMarketPrice,
		Change:    oldQ.RegularMarketChange,
		ChangePct: oldQ.RegularMarketChangePercent,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *YahooDataService) SearchMarketNews(ctx context.Context, query string) ([]NewsItem, error) {
	// RSS Feeds
	feeds := map[string]string{
		"us":  "https://feeds.content.dowjones.io/public/rss/mw_topstories", // Fallback to MarketWatch
		"all": "https://feeds.content.dowjones.io/public/rss/mw_topstories",

		// China: Google News China (Very stable)
		"cn": "https://news.google.com/rss/search?q=A%E8%82%A1&hl=zh-CN&gl=CN&ceid=CN:zh-CN",

		// Macro: CNBC & WSJ
		"macro":    "https://search.cnbc.com/rs/search/combinedcms/view.xml?partnerId=wrss01&id=20910258",
		"us_macro": "https://search.cnbc.com/rs/search/combinedcms/view.xml?partnerId=wrss01&id=20910258",
		"wsj_econ": "https://feeds.a.dj.com/rss/WSJcomUSBusiness.xml",

		// Crypto: Cointelegraph
		"crypto": "https://cointelegraph.com/rss",

		// Others (Removed broken/paid feeds like Reuters/Bloomberg/TheBlock)
		"reuters":   "https://feeds.content.dowjones.io/public/rss/mw_topstories",                                             // Fallback to MarketWatch
		"bloomberg": "https://search.cnbc.com/rs/search/combinedcms/view.xml?partnerId=wrss01&id=10000664",                    // CNBC Top News
		"theblock":  "https://cointelegraph.com/rss",                                                                          // Fallback to Cointelegraph
		"panews":    "https://news.google.com/rss/search?q=%E5%8A%A0%E5%AF%86%E8%B4%A7%E5%B8%81&hl=zh-CN&gl=CN&ceid=CN:zh-CN", // Fallback to Google News Crypto
	}

	urlStr, ok := feeds[query]
	if !ok {
		// If query is not a predefined category, use Google News Search
		// Encode query
		encodedQuery := url.QueryEscape(query)
		// Default to English if not containing Chinese, else Chinese
		lang := "en-US"
		region := "US"
		ceid := "US:en"

		// Simple check for Chinese characters
		for _, r := range query {
			if r >= 0x4e00 && r <= 0x9fff {
				lang = "zh-CN"
				region = "CN"
				ceid = "CN:zh-CN"
				break
			}
		}

		urlStr = fmt.Sprintf("https://news.google.com/rss/search?q=%s&hl=%s&gl=%s&ceid=%s", encodedQuery, lang, region, ceid)
	}

	fp := gofeed.NewParser()
	// Set User-Agent for gofeed
	fp.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	feed, err := fp.ParseURL(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch news: %v", err)
	}

	var news []NewsItem
	for i, item := range feed.Items {
		if i >= 5 {
			break
		}

		summary := item.Description
		if len(summary) > 200 {
			summary = summary[:200] + "..."
		}

		news = append(news, NewsItem{
			Title:   item.Title,
			Summary: summary,
			Source:  feed.Title,
			Time:    item.Published,
		})
	}

	return news, nil
}

func (s *YahooDataService) GetMarketIndex(ctx context.Context) ([]IndexQuote, error) {
	// Re-implement GetMarketIndex using new getYahooPriceV8 logic to avoid finance-go issues
	symbols := []string{"^GSPC", "^IXIC", "^HSI", "000001.SS", "BTC-USD", "GC=F"}

	var indices []IndexQuote
	for _, sym := range symbols {
		q, err := s.GetMarketQuote(ctx, sym)
		if err == nil {
			indices = append(indices, IndexQuote{
				Name:      sym, // Ideally map to shortname
				Value:     q.Price,
				Change:    q.Change,
				ChangePct: q.ChangePct,
			})
		}
	}
	return indices, nil
}

func (s *YahooDataService) GetSecurityAnalysis(ctx context.Context, symbol string, assetType string) (*SecurityAnalysis, error) {
	symbol = normalizeSymbol(symbol)

	// 1. Get Real Quote (Using V8)
	q, err := s.GetMarketQuote(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote for %s: %v", symbol, err)
	}

	// 2. Get Historical Data (K-Lines) using GetHistoricalQuotes (V8 API)
	// Get last 90 days for technical analysis
	klines, err := s.GetHistoricalQuotes(ctx, symbol, "1d", "3mo")
	if err != nil || len(klines) == 0 {
		// If failed, return basic info
		return &SecurityAnalysis{
			Symbol:       symbol,
			AssetType:    assetType,
			CurrentPrice: q.Price,
			Trend:        "unknown",
		}, nil
	}

	var closes []float64
	for _, k := range klines {
		closes = append(closes, k.Close)
	}

	// 3. Calculate Indicators
	ma20 := calculateSMA(closes, 20)
	ma60 := calculateSMA(closes, 60)
	rsi := calculateRSI(closes, 14)

	// Calculate Volume Ratio (Last Volume / MA5 Volume)
	volRatio := 0.0
	if len(klines) >= 6 {
		lastVol := klines[len(klines)-1].Volume
		sumVol := 0.0
		for _, k := range klines[len(klines)-6 : len(klines)-1] {
			sumVol += k.Volume
		}
		avgVol := sumVol / 5.0
		if avgVol > 0 {
			volRatio = lastVol / avgVol
		}
	}

	trend := "sideways"
	if ma20 > 0 && ma60 > 0 {
		if ma20 > ma60 && q.Price > ma20 {
			trend = "bullish"
		} else if ma20 < ma60 && q.Price < ma20 {
			trend = "bearish"
		}
	}

	recentKLines := klines
	if len(klines) > 5 {
		recentKLines = klines[len(klines)-5:]
	}

	// Simple Support/Resistance based on recent high/low
	support := q.Price
	resistance := q.Price
	if len(klines) > 0 {
		// Default to recent low/high
		support = klines[len(klines)-1].Close
		resistance = klines[len(klines)-1].Close

		// Look back 20 periods or max available
		lookback := 20
		if len(klines) < 20 {
			lookback = len(klines)
		}

		lows := 1000000.0
		highs := 0.0
		for _, k := range klines[len(klines)-lookback:] {
			if k.Close < lows {
				lows = k.Close
			}
			if k.Close > highs {
				highs = k.Close
			}
		}
		support = lows
		resistance = highs
	}

	return &SecurityAnalysis{
		Symbol:          symbol,
		AssetType:       assetType,
		CurrentPrice:    q.Price,
		MA20:            ma20,
		MA60:            ma60,
		RSI:             rsi,
		VolumeRatio:     volRatio,
		Trend:           trend,
		SupportLevel:    support,
		ResistanceLevel: resistance,
		RecentKLines:    recentKLines,
	}, nil
}

func (s *YahooDataService) GetHistoricalQuotes(ctx context.Context, symbol string, interval string, rangeStr string) ([]KLineItem, error) {
	symbol = normalizeSymbol(symbol)
	// Yahoo API: https://query1.finance.yahoo.com/v8/finance/chart/{symbol}?interval={interval}&range={range}
	apiURL := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=%s&range=%s", symbol, interval, rangeStr)

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("yahoo chart api error: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var chartResp YahooChartResponse
	if err := json.Unmarshal(body, &chartResp); err != nil {
		return nil, err
	}

	if len(chartResp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no historical data found for %s", symbol)
	}

	result := chartResp.Chart.Result[0]
	timestamps := result.Timestamp

	// Check if Quotes are available
	if len(result.Indicators.Quote) == 0 {
		return nil, fmt.Errorf("no quote data found for %s", symbol)
	}
	quote := result.Indicators.Quote[0]

	var klines []KLineItem
	for i, ts := range timestamps {
		if i >= len(quote.Close) {
			break
		}
		// Skip null values
		if quote.Close[i] == 0 && quote.Volume[i] == 0 {
			continue
		}

		klines = append(klines, KLineItem{
			Date:   time.Unix(ts, 0).Format("2006-01-02"),
			Close:  quote.Close[i],
			Volume: float64(quote.Volume[i]),
		})
	}

	return klines, nil
}

type FearGreedResponse struct {
	Name string `json:"name"`
	Data []struct {
		Value               string `json:"value"`
		ValueClassification string `json:"value_classification"`
		Timestamp           string `json:"timestamp"`
	} `json:"data"`
}

func (s *YahooDataService) GetMarketSentiment(ctx context.Context, market string) (*SentimentData, error) {
	if market == "crypto" {
		resp, err := http.Get("https://api.alternative.me/fng/")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("fear & greed api error: %d", resp.StatusCode)
		}

		var result FearGreedResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		if len(result.Data) > 0 {
			score, _ := strconv.ParseFloat(result.Data[0].Value, 64)
			ts, _ := strconv.ParseInt(result.Data[0].Timestamp, 10, 64)
			return &SentimentData{
				Market:      "crypto",
				Score:       score,
				Label:       result.Data[0].ValueClassification,
				Description: fmt.Sprintf("Crypto Fear & Greed Index is %s", result.Data[0].ValueClassification),
				Timestamp:   ts,
			}, nil
		}
	}

	// Fallback or Stock Sentiment (Simple Heuristic using SP500)
	// If market is not crypto, we try to guess based on SP500 trend
	q, err := s.GetMarketQuote(ctx, "^GSPC")
	if err == nil {
		// Use ChangePct as a very rough daily sentiment
		label := "Neutral"
		if q.ChangePct > 1.0 {
			label = "Greed"
		} else if q.ChangePct < -1.0 {
			label = "Fear"
		}

		return &SentimentData{
			Market:      "us_stock",
			Score:       50 + q.ChangePct*10, // Rough mapping
			Label:       label,
			Description: fmt.Sprintf("S&P 500 Daily Change is %.2f%%", q.ChangePct),
			Timestamp:   time.Now().Unix(),
		}, nil
	}

	return nil, fmt.Errorf("sentiment data not available for %s", market)
}

// Helpers (Same as before)
func calculateSMA(data []float64, period int) float64 {
	if len(data) < period {
		return 0
	}
	sum := 0.0
	for _, v := range data[len(data)-period:] {
		sum += v
	}
	return sum / float64(period)
}

func calculateRSI(data []float64, period int) float64 {
	if len(data) < period+1 {
		return 50 // default
	}

	gains := 0.0
	losses := 0.0

	// First period
	for i := len(data) - period; i < len(data); i++ {
		change := data[i] - data[i-1]
		if change > 0 {
			gains += change
		} else {
			losses -= change
		}
	}

	if losses == 0 {
		return 100
	}

	rs := gains / losses
	return 100 - (100 / (1 + rs))
}
