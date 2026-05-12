# Stock Daily — A股日线价差分析工具

基于 Tushare API 的中国 A 股日线数据获取与价差统计分析 CLI 工具。

## 功能

- **数据获取**: 从 Tushare 拉取日线 OHLCV 数据，支持增量更新
- **价差计算**: 自动计算 6 种价差（高-开、开-低、高-低、开-收、高-收、低-收）
- **统计分析**: 多时间窗口（全部/近90日/近30日/近15日）的均值、中位数、众数
- **分布展示**: 直方图式分布表，多个时间窗口并排对比
- **终端友好**: CJK 字符对齐，中文表头正确显示

## 快速开始

### 1. 安装依赖

```bash
python -m venv .venv
source .venv/bin/activate  # Windows: .venv\Scripts\activate
pip install -r requirements.txt
```

### 2. 配置 Tushare Token

复制 `.env.example` 为 `.env`，填入你的 Tushare token：

```bash
cp .env.example .env
```

```env
TUSHARE_TOKEN=你的tushare_token
```

Token 申请地址: [https://tushare.pro/register](https://tushare.pro/register)

### 3. 获取数据

```bash
# 增量获取（从上次记录日期开始）
python stock.py fetch --stocks 603778,000890

# 指定天数
python stock.py fetch --stocks 603778 --days 90

# 全量历史（回溯约10年）
python stock.py fetch --stocks 603778 --days all
```

### 4. 查看分析

```bash
# 默认显示近30日数据 + 价差分析
python stock.py show --stock 603778

# 指定日期范围
python stock.py show --stock 603778 --from 2025-04-01 --to 2025-05-12

# 显示全部6种价差
python stock.py show --stock 603778 --analyze --all
```

## 价差说明

| 价差 | 计算 | 用途 |
|------|------|------|
| 高-开 (spread_oh) | 最高价 − 开盘价 | 高抛目标参考 |
| 开-低 (spread_ol) | 开盘价 − 最低价 | 低吸目标参考 |
| 高-低 (spread_hl) | 最高价 − 最低价 | 全日振幅，衡量波动性 |
| 开-收 (spread_oc) | 开盘价 − 收盘价 | 日内方向，正=收阳，负=收阴 |
| 高-收 (spread_hc) | 最高价 − 收盘价 | 盘中高点回落幅度 |
| 低-收 (spread_lc) | 最低价 − 收盘价 | 盘中低点反弹幅度 |

默认只显示高-开和开-低两种，使用 `--all` 查看全部。

## 支持的股票代码

| 代码前缀 | 市场 | 板块 |
|----------|------|------|
| 600/601/603/605 | 上交所 | 主板 |
| 688 | 上交所 | 科创板 |
| 000/001 | 深交所 | 主板 |
| 002 | 深交所 | 中小板 |
| 300 | 深交所 | 创业板 |
| 510/511/512/515 | 上交所 | ETF |
| 159 | 深交所 | ETF |

直接使用 6 位数字代码即可，工具自动识别交易所。

## 项目结构

```
stock/
├── stock.py          # CLI 入口，数据展示
├── config.py         # 配置，股票代码转换
├── fetcher.py        # Tushare API 数据获取
├── db.py             # SQLite 数据库操作
├── analysis.py       # 统计分析与分布计算
├── requirements.txt  # Python 依赖
├── .env.example      # 环境变量模板
├── data/             # SQLite 数据库（git 忽略）
└── tests/            # 测试用例
```

## 测试

```bash
python -m pytest tests/ -v
```

## 依赖

- Python 3.10+
- tushare — 数据源
- pandas — 数据处理
- python-dotenv — 环境变量
- tabulate — 表格格式化
- pytest — 测试框架
