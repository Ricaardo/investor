package main

import (
	"context"
	"fmt"
	"time"

	"investor/internal/dataservice"

	"github.com/mmcdole/gofeed"
)

func main() {
	fmt.Println("🔍 Starting Data Source Health Check...")
	fmt.Println("----------------------------------------")

	ds := dataservice.NewYahooDataService()
	ctx := context.Background()

	// 1. Check Yahoo Finance Quote (V8 Chart API)
	checkQuote(ctx, ds, "AAPL (US)", "AAPL")
	checkQuote(ctx, ds, "Tencent (HK)", "0700.HK")
	checkQuote(ctx, ds, "Moutai (CN)", "600519.SS")
	checkQuote(ctx, ds, "Gold (Futures)", "GC=F")

	// 2. Check Binance Quote
	checkQuote(ctx, ds, "Bitcoin (OKX)", "BTCUSDT") // This will now try OKX first

	// 2.1 Check Typo Correction
	checkQuote(ctx, ds, "Typo: APPL", "APPL") // Should resolve to AAPL or fail if not aliased
	checkQuote(ctx, ds, "Futures: Copper", "铜")
	checkQuote(ctx, ds, "Futures: Soybean", "大豆")
	checkQuote(ctx, ds, "Futures: 10Y Bond", "10年美债")
	checkQuote(ctx, ds, "Futures: Euro", "欧元期货")
	checkQuote(ctx, ds, "Forex: USD/CNY", "人民币")
	checkQuote(ctx, ds, "Forex: USD/JPY", "日元")
	checkQuote(ctx, ds, "Index: DXY", "美元指数")

	// 3. Check Yahoo Search API
	fmt.Printf("\n[Search API] Testing '腾讯'...\n")
	q, err := ds.GetMarketQuote(ctx, "腾讯")
	if err == nil && q.Symbol == "0700.HK" {
		fmt.Printf("✅ PASS: '腾讯' -> %s\n", q.Symbol)
	} else {
		fmt.Printf("❌ FAIL: '腾讯' -> %v (err: %v)\n", q, err)
	}

	// 4. Check RSS Feeds (Updated List)
	fmt.Println("\n[RSS Feeds] Checking availability...")
	feeds := map[string]string{
		"US (MarketWatch)":       "https://feeds.content.dowjones.io/public/rss/mw_topstories",
		"CN (Google News)":       "https://news.google.com/rss/search?q=A%E8%82%A1&hl=zh-CN&gl=CN&ceid=CN:zh-CN",
		"Macro (CNBC)":           "https://search.cnbc.com/rs/search/combinedcms/view.xml?partnerId=wrss01&id=20910258",
		"Macro (WSJ)":            "https://feeds.a.dj.com/rss/WSJcomUSBusiness.xml",
		"Crypto (Cointelegraph)": "https://cointelegraph.com/rss",
	}

	fp := gofeed.NewParser()
	// Set User-Agent to match production service
	fp.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	for name, url := range feeds {
		checkRSS(fp, name, url)
	}

	// 5. Check Historical Quotes
	fmt.Println("\n[Historical Data] Checking availability...")
	checkHistory(ctx, ds, "AAPL (Last 1mo)", "AAPL", "1d", "1mo")
	checkHistory(ctx, ds, "BTC (Last 5d)", "BTC-USD", "1h", "5d")

	// 6. Check Market Sentiment
	fmt.Println("\n[Sentiment] Checking availability...")
	checkSentiment(ctx, ds, "Crypto", "crypto")
	checkSentiment(ctx, ds, "US Stock", "us_stock")

	fmt.Println("----------------------------------------")
	fmt.Println("✅ Health Check Completed.")
}

func checkQuote(ctx context.Context, ds dataservice.DataService, name, symbol string) {
	start := time.Now()
	q, err := ds.GetMarketQuote(ctx, symbol)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ FAIL: %-20s (%s) - Error: %v\n", name, symbol, err)
	} else {
		fmt.Printf("✅ PASS: %-20s (%s) - Price: %.2f Change: %.2f%% (took %v)\n", name, symbol, q.Price, q.ChangePct, duration)
	}
}

func checkHistory(ctx context.Context, ds dataservice.DataService, name, symbol, interval, rng string) {
	start := time.Now()
	klines, err := ds.GetHistoricalQuotes(ctx, symbol, interval, rng)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ FAIL: %-20s (%s) - Error: %v\n", name, symbol, err)
	} else {
		fmt.Printf("✅ PASS: %-20s (%s) - Got %d bars (took %v)\n", name, symbol, len(klines), duration)
		if len(klines) > 0 {
			last := klines[len(klines)-1]
			fmt.Printf("        Last Bar: %s Close: %.2f\n", last.Date, last.Close)
		}
	}
}

func checkSentiment(ctx context.Context, ds dataservice.DataService, name, market string) {
	start := time.Now()
	sent, err := ds.GetMarketSentiment(ctx, market)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ FAIL: %-20s - Error: %v\n", name, err)
	} else {
		fmt.Printf("✅ PASS: %-20s - Score: %.0f (%s) (took %v)\n", name, sent.Score, sent.Label, duration)
	}
}

func checkRSS(fp *gofeed.Parser, name, url string) {
	start := time.Now()

	// Set 10s timeout for RSS
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	feed, err := fp.ParseURLWithContext(url, ctx)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ FAIL: %-25s - Error: %v\n", name, err)
	} else if len(feed.Items) == 0 {
		fmt.Printf("⚠️ WARN: %-25s - OK (HTTP 200) but 0 items\n", name)
	} else {
		fmt.Printf("✅ PASS: %-25s - OK (%d items, took %v)\n", name, len(feed.Items), duration)
	}
}
