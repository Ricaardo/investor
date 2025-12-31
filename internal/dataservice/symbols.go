package dataservice

import (
	"fmt"
	"regexp"
	"strings"
)

// normalizeSymbol helps guess suffix for A-shares or map common names
func normalizeSymbol(symbol string) string {
	lowerSym := strings.ToLower(symbol)

	// 1. Common Alias Map
	// Moved from yahoo.go to clean up code
	aliasMap := map[string]string{
		"比特币": "BTCUSDT", "btc": "BTCUSDT",
		"以太坊": "ETHUSDT", "eth": "ETHUSDT",
		"sol": "SOLUSDT",
		"bnb": "BNBUSDT",
		"黄金":  "GC=F", "gold": "GC=F",
		"白银": "SI=F", "silver": "SI=F",
		"原油": "CL=F", "oil": "CL=F",
		"纳指": "^IXIC", "nasdaq": "^IXIC",
		"标普": "^GSPC", "sp500": "^GSPC",
		"恒指": "^HSI", "hsi": "^HSI",
		"上证": "000001.SS", "shanghai": "000001.SS",
		"腾讯": "0700.HK", "tencent": "0700.HK",
		"阿里": "BABA", "alibaba": "BABA",
		"特斯拉": "TSLA", "tesla": "TSLA",
		"苹果": "AAPL", "apple": "AAPL",
		"英伟达": "NVDA", "nvidia": "NVDA",
		"微软": "MSFT", "microsoft": "MSFT",
		"谷歌": "GOOG", "google": "GOOG",
		"亚马逊": "AMZN", "amazon": "AMZN",
		"appl": "AAPL", // Common typo

		// --- Energy Futures ---
		"wti": "CL=F",
		"布伦特": "BZ=F", "brent": "BZ=F",
		"天然气": "NG=F", "natgas": "NG=F",
		"燃油": "HO=F", "heating_oil": "HO=F",
		"汽油": "RB=F", "gasoline": "RB=F",

		// --- Metal Futures ---
		"铜": "HG=F", "copper": "HG=F",
		"铂金": "PL=F", "platinum": "PL=F",
		"钯金": "PA=F", "palladium": "PA=F",
		"铝": "ALI=F", "aluminum": "ALI=F",

		// --- Agriculture Futures ---
		"玉米": "ZC=F", "corn": "ZC=F",
		"大豆": "ZS=F", "soybean": "ZS=F", "soybeans": "ZS=F",
		"豆油": "ZL=F", "soybean_oil": "ZL=F",
		"豆粕": "ZM=F", "soybean_meal": "ZM=F",
		"小麦": "ZW=F", "wheat": "ZW=F",
		"糖": "SB=F", "sugar": "SB=F",
		"咖啡": "KC=F", "coffee": "KC=F",
		"可可": "CC=F", "cocoa": "CC=F",
		"棉花": "CT=F", "cotton": "CT=F",
		"活牛": "LE=F", "live_cattle": "LE=F",
		"瘦肉猪": "HE=F", "lean_hogs": "HE=F",

		// --- Index Futures ---
		"标普期货": "ES=F", "es": "ES=F",
		"纳指期货": "NQ=F", "nq": "NQ=F",
		"道指期货": "YM=F", "ym": "YM=F",
		"罗素期货": "RTY=F", "rty": "RTY=F",
		"恐慌指数期货": "VX=F", "vix_future": "VX=F",

		// --- Bond Futures ---
		"10年美债": "ZN=F", "10y_bond": "ZN=F",
		"30年美债": "ZB=F", "30y_bond": "ZB=F",
		"5年美债": "ZF=F", "5y_bond": "ZF=F",
		"2年美债": "ZT=F", "2y_bond": "ZT=F",

		// --- Currency Futures ---
		"欧元期货": "6E=F", "eur_future": "6E=F",
		"日元期货": "6J=F", "jpy_future": "6J=F",
		"英镑期货": "6B=F", "gbp_future": "6B=F",
		"澳元期货": "6A=F", "aud_future": "6A=F",

		// --- Forex (Spot) ---
		// Majors
		"欧元": "EURUSD=X", "eur": "EURUSD=X", "eurusd": "EURUSD=X",
		"日元": "JPY=X", "jpy": "JPY=X", "usdjpy": "JPY=X", // Yahoo format for USD/JPY is JPY=X
		"英镑": "GBPUSD=X", "gbp": "GBPUSD=X", "gbpusd": "GBPUSD=X",
		"澳元": "AUDUSD=X", "aud": "AUDUSD=X", "audusd": "AUDUSD=X",
		"加元": "CAD=X", "cad": "CAD=X", "usdcad": "CAD=X",
		"瑞郎": "CAD=X", "chf": "CHF=X", "usdchf": "CHF=X", // Typo in original line? CAD=X? Fixing to CHF=X
		"纽元": "NZDUSD=X", "nzd": "NZDUSD=X", "nzdusd": "NZDUSD=X",

		// Crosses & Exotics
		"人民币": "CNY=X", "cny": "CNY=X", "usdcny": "CNY=X",
		"离岸人民币": "CNH=X", "cnh": "CNH=X", "usdcnh": "CNH=X",
		"港币": "HKD=X", "hkd": "HKD=X", "usdhkd": "HKD=X",
		"台币": "TWD=X", "twd": "TWD=X", "usdtwd": "TWD=X",
		"韩元": "TWD=X", "krw": "KRW=X", "usdkrw": "KRW=X", // Typo TWD? Fixing to KRW=X
		"新加坡元": "SGD=X", "sgd": "SGD=X", "usdsgd": "SGD=X",
		"卢布": "RUB=X", "rub": "RUB=X", "usdrub": "RUB=X",
		"卢比": "INR=X", "inr": "INR=X", "usdinr": "INR=X",
		"泰铢": "THB=X", "thb": "THB=X", "usdthb": "THB=X",
		"越南盾": "THB=X", "vnd": "VND=X", "usdvnd": "VND=X", // Typo THB? Fixing to VND=X
		"巴西雷亚尔": "BRL=X", "brl": "BRL=X", "usdbrl": "BRL=X",
		"南非兰特": "ZAR=X", "zar": "ZAR=X", "usdzar": "ZAR=X",
		"土耳其里拉": "TRY=X", "try": "TRY=X", "usdtry": "TRY=X",
		"墨西哥比索": "MXN=X", "mxn": "MXN=X", "usdmxn": "MXN=X",

		// Index
		"美元指数": "DX-Y.NYB", "dxy": "DX-Y.NYB", "usd_index": "DX-Y.NYB",
	}
	if val, ok := aliasMap[lowerSym]; ok {
		return val
	}

	// 2. Local StockMap
	if val, ok := StockMap[symbol]; ok {
		return val
	}
	if len(symbol) > 3 {
		for name, code := range StockMap {
			if strings.Contains(name, symbol) || strings.Contains(symbol, name) {
				return code
			}
		}
	}

	// 3. Auto Suffix
	if matched, _ := regexp.MatchString(`^\d{6}$`, symbol); matched {
		if symbol[0] == '6' {
			return symbol + ".SS"
		} else if symbol[0] == '0' || symbol[0] == '3' {
			return symbol + ".SZ"
		}
	}
	if matched, _ := regexp.MatchString(`^\d{4}$`, symbol); matched {
		return fmt.Sprintf("%04s.HK", symbol)
	}

	// 4. Yahoo Online Search
	isTicker := regexp.MustCompile(`^[A-Z0-9\-\.=]+$`).MatchString(strings.ToUpper(symbol))
	if !isTicker || len(symbol) > 5 {
		found := searchYahooSymbol(symbol)
		if found != "" {
			return found
		}
	}

	return symbol
}
