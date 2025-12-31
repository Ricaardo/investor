# Investor 使用文档

Investor 是一个基于大语言模型（LLM）的**全资产智能投资分析中台**。它不仅仅是一个行情查询工具，更是一个具备机构级思维框架的 AI 投资顾问。

## 🌟 核心能力

Investor 旨在辅助专业投资者进行决策，核心能力包括：

1.  **全资产行情覆盖**: 支持 **股票 (A股/港美股)**、**加密货币 (Crypto)**、**外汇 (Forex)**、**大宗商品 (Commodity)** 和 **宏观指数**。
2.  **深度技术分析**: 自动计算 MA, RSI, MACD, 支撑/压力位，并识别技术形态（如背离、超买/超卖）。
3.  **基本面与宏观洞察**: 结合实时新闻搜索，分析美联储政策、财报数据、链上资金流向。
4.  **情景推演与风控**: 提供乐观/悲观剧本推演，计算盈亏比，并进行“批判性思考”以规避盲点。
5.  **多渠道接入**: 支持 **飞书 (Feishu)** 机器人对话，并提供 **REST API** 供 Coze/Dify 等第三方平台调用。

---

## 🚀 快速开始

### 1. 环境准备
确保已安装 Go 1.21+。

### 2. 配置
在项目根目录创建 `.env` 文件（参考 `.env.example`）：

```bash
# LLM 配置 (支持 OpenAI/DeepSeek 等兼容接口)
LLM_PROVIDER=openai
LLM_API_KEY=your_api_key
LLM_API_URL=https://api.openai.com/v1

# 飞书配置 (可选)
FEISHU_APP_ID=cli_xxx
FEISHU_APP_SECRET=xxx
FEISHU_ENCRYPT_KEY=xxx
FEISHU_VERIFICATION_TOKEN=xxx

# 服务端口
PORT=8080
```

### 3. 启动服务
```bash
go run cmd/server/main.go
```
服务启动后：
- **REST API**: 监听 `http://localhost:8080`
- **飞书 Bot**: 自动建立 WebSocket 连接

---

## 💬 交互指令指南

Investor AI 经过深度调优，支持多种复杂的自然语言指令。以下是典型的使用场景：

### 1. 深度个股/标的分析
> **指令示例**: “分析英伟达” / “分析 BTC” / “分析 黄金”

AI 将返回一份机构级研报，包含：
- **核心观点**: 一句话总结（如“缩量盘整”）。
- **适用周期**: 短线/中线/长线。
- **关键数据**: 现价、涨跌幅、均线位置、RSI 情绪。
- **深度逻辑**: 宏观面 + 技术面 + 资金面（机构博弈）。
- **盈亏比**: 明确的阻力位（目标）和支撑位（止损）。
- **情景推演**: 📈 乐观剧本 vs 📉 悲观剧本。

### 2. 市场对比分析
> **指令示例**: “对比一下 BTC 和 ETH” / “腾讯和阿里哪个基本面更好？”

AI 会自动生成 **Markdown 表格**，横向对比两者的核心指标（如今年涨幅、波动率、市值、PE/估值等），并给出强弱判断。

### 3. 宏观与新闻解读
> **指令示例**: “搜索最近关于美联储降息的新闻” / “为什么今天原油大跌？”

AI 会调用实时搜索工具，汇总多方新闻源（Bloomberg, Reuters, Coindesk），提炼出导致市场波动的核心驱动力，而非简单的复读标题。

### 4. 市场情绪与策略
> **指令示例**: “现在的市场情绪怎么样？” / “恐慌指数是多少？”

AI 会查询恐慌与贪婪指数（Fear & Greed Index），结合技术指标，判断当前是“极度恐慌（可能的买点）”还是“极度贪婪（风险积聚）”。

### 5. 新手/专家模式切换
> **指令示例**: “什么是 RSI 指标？”（自动触发教学模式）

- **专家模式 (默认)**: 使用“背离”、“流动性猎取”、“Gamma Squeeze”等专业术语，简练直接。
- **新手模式**: 当检测到基础问题时，会自动解释术语，循循善诱。

---

## 🔌 开发者接口 (API)

您可以将 Investor 集成到自己的工作流（如 Notion, Obsidian, Coze）中。

**接口地址**: `POST /api/v1/chat`

**请求示例**:
```json
{
    "user_id": "test_user",
    "text": "分析特斯拉，给出短线交易计划",
    "platform": "api"
}
```

**响应示例**:
```json
{
    "response": "🎯 **核心观点**: 处于三角形收敛末端，等待方向选择...\n..."
}
```

---

## 🛠 扩展与自定义

### 添加自定义数据源
系统采用注册表模式，您可以轻松接入内部数据（如 Bloomberg 终端、私有数据库）。

1. 在 `internal/dataservice/` 下实现 `DataService` 接口。
2. 在 `cmd/server/main.go` 中注册：
   ```go
   mySource := dataservice.NewMyPrivateSource()
   dataservice.GetRegistry().Register("my_source", mySource)
   ```

### 接入新渠道
参考 `internal/adapter/rest` 实现新的 Adapter（如钉钉、Telegram），并在 `main.go` 中启动即可。

---

## ⚠️ 免责声明
Investor 提供的所有分析基于历史数据和概率模型，**仅供参考，不构成投资建议**。市场有风险，投资需谨慎。
