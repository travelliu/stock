# 前端重构 & 字段统一 — Design Spec

**Date:** 2026-05-14
**Status:** Draft(待用户审查)
**Author:** travelliu(with Claude)
**Reference UI:** `/root/code/gitlab/mogdb_en/mtk/pkg/mtkd/web/`
**Prior spec:** [`2026-05-14-go-vue-rewrite-design.md`](2026-05-14-go-vue-rewrite-design.md)、[`2026-05-14-pkg-restructure-design.md`](2026-05-14-pkg-restructure-design.md)

---

## 0. 背景与目标

Python → Go + Vue 重写(prior spec)与 pkg 重组完成后,当前 `web/` 暴露三类问题:

1. **UI 简陋**:200px 浅色侧栏 + 单层菜单,无 SCSS、无国际化、无图标系统,与 `mtkd/web`(65px 暗色侧栏 + iconfont + i18n + `g-*` 通用组件)风格落差大。
2. **字段不一致**:
   - 后端 `pkg/models` 多数字段已 camelCase,但残留 `userName`(应为 `username`)、`userID`(应为 `userId`)、`DailyBar.Open` 有重复 JSON tag bug。
   - 后端 `http/auth.go::loginReq`、`http/me.go` 多个 Req/Resp 是 inline struct,未被复用。
   - 前端 `stores/auth.ts`(`tushare_token`)、`stores/portfolio.ts`(`ts_code`/`added_at`)、`components/AnalysisPanel.vue`(`model_table`)使用 snake_case,与后端实际返回的 camelCase 不匹配,实际取不到字段。
   - CLI `cli/cmd/portfolio.go` 用 `[]*models.PortfolioReq` 解析 GET `/api/portfolio`(应为 `[]*models.Portfolio`);`cli/cmd/login.go` 内联 `username` 字段(后端是 `userName`)。
3. **DTO 散落**:后端 inline Req、CLI inline 解析、前端单独 TS interface 各自维护,无单一真相源。

**目标**:
- 重写 `web/src`,贴合 mtkd 风格(暗色窄侧栏、SCSS 变量化、i18n、`g-*` 通用组件、`apis/` 拆分)。
- 所有 HTTP Req/Resp DTO 统一收敛至 `pkg/models/`,handler 与 CLI 共用同一份。
- 前后端 JSON 字段统一为 **标准小驼峰**(`userId` 而非 `userID`)。
- 前端引入 `vue-i18n` 框架,zh-CN 文案完整、en-US 占位以便后续补全。

**非目标**:
- 不增加新业务功能(登录/持仓/分析/草稿/历史/用户管理/同步保持现有能力)。
- 不引入 echarts、不接入外部 iconfont CDN、不引入 OpenAPI 自动生成。

---

## 1. 关键决策记录

| 决策点 | 选定方案 | 备注 |
|---|---|---|
| 范围 | UI 美化 + 全栈字段统一 + i18n 框架 | en-US 仅留 key,fallback 到 zh |
| DTO 共享 | Go inline 搬到 `pkg/models`;前端手写 `src/types/api.ts` | 不引入代码生成器 |
| 详情页结构 | 2 Tab(基础信息+价差 / 详细统计),贴合 `前端设计文档.md` | 现有 3 Tab(分析/历史/草稿)拆并 |
| 配色 | 涨红跌绿 + Element Plus 蓝主色 | 不覆盖 `--el-color-primary` |
| 字段规范 | 标准 camelCase(`username` / `userId` / `stockId`) | 不再使用 `userName` / `userID` 这种全大写缩写 |
| 实施路径 | 一次性重写 `web/src`(保留 vite/tsconfig/playwright 配置) | feature branch + e2e + smoke 后合 master |

---

## 2. 后端 + CLI:契约层重构

### 2.1 `pkg/models` 重组

新增文件,从 handler/cli 把 inline struct 搬过来:

**`pkg/models/auth.go`**(新增)
```go
type LoginReq struct {
    Username string `json:"username"`
    Password string `json:"password"`
}
```

**`pkg/models/me.go`**(新增)
```go
type ChangePasswordReq struct {
    Old string `json:"old"`
    New string `json:"new"`
}

type SetTushareTokenReq struct {
    Token string `json:"token"`
}

type IssueTokenReq struct {
    Name      string     `json:"name"`
    ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

type IssueTokenResp struct {
    Token    string    `json:"token"`
    Metadata *ApiToken `json:"metadata"`
}
```

**`pkg/models/portfolio.go`** 不变,但 CLI 必须用 `Portfolio`(GET 返回类型)而非 `PortfolioReq`(POST/PUT 入参类型)。

### 2.2 字段重命名

| 模型 | 字段 | 旧 JSON | 新 JSON |
|---|---|---|---|
| `User` | `Username` | `userName` | **`username`** |
| `Portfolio` | `UserID` | `userID` | **`userId`** |
| `DailyBar` | `Open` | 重复 tag bug `json:"open" json:"open,omitempty"` | **`open,omitempty`**(单次) |

通用规则:
- ID 字段命名为 `id` / `userId` / `stockId`,不使用 `userID` / `stockID` 全大写缩写。
- 主键、`createdAt`、`updatedAt`、`addedAt` 等系统时间字段去掉 `omitempty`,保证总是返回。
- 现有已是标准 camelCase 的字段(`tsCode` / `tushareToken` / `tradeDate` 等)不动。

### 2.3 Inline → models 迁移清单

| 文件 | 原 inline 名 | 替换为 |
|---|---|---|
| `pkg/stockd/http/auth.go::Login` | `loginReq` | `models.LoginReq` |
| `pkg/stockd/http/me.go::IssueToken` | inline req | `models.IssueTokenReq` |
| `pkg/stockd/http/me.go::IssueToken` | `gin.H{"token","metadata"}` | `models.IssueTokenResp` |
| `pkg/stockd/http/me.go::SetTushareToken` | inline req | `models.SetTushareTokenReq` |
| `pkg/stockd/http/me.go::ChangePassword` | inline req | `models.ChangePasswordReq` |
| `pkg/cli/cmd/login.go` | inline me struct | `models.User` |
| `pkg/cli/cmd/portfolio.go` GET | `[]*models.PortfolioReq` | `[]*models.Portfolio` |

迁移后所有 handler 函数体的 `var req inline{...}` 都改为 `var req models.XxxReq`,CLI 同步更新。

### 2.4 后端测试

- `pkg/models/models_test.go` 增 `TestModelJSONFields` table-driven:每个 struct marshal 后字段名 == 期望集合(防止字段名漂移)。
- `pkg/stockd/http/auth_test.go`、`pkg/stockd/http/me_test.go` 更新断言:用 `models.LoginReq` 等构造请求,断言响应字段名。
- `pkg/cli/client/client_test.go` 补 portfolio 字段对齐用例:mock server 返回 `Portfolio` 完整字段,CLI 解析后保留 ID/AddedAt。

### 2.5 文档提示

`pkg/models/doc.go` 添加包级注释:

```go
// Package models is the single source of truth for all HTTP request/response DTOs.
// The frontend type definitions in web/src/types/api.ts MUST stay in sync;
// when you change a field here, update both sides in the same PR.
package models
```

---

## 3. 前端架构(`web/`)

### 3.1 目录结构

```
web/src/
├── App.vue                    # el-config-provider + layout(侧栏 + main)
├── main.ts                    # 入口:ElementPlus + Pinia + Router + i18n + 全局组件
├── env.d.ts
├── assets/
│   ├── css/
│   │   ├── index.scss         # body 重置、g-* 公共类、:root 变量(涨红跌绿)
│   │   ├── element-reset.scss # 卡片圆角 4px、按钮高度等
│   │   └── mixin.scss         # @mixin scrollbar()
│   └── image/
│       └── logo-mini.svg      # 简单 SVG logo
├── types/
│   └── api.ts                 # 唯一手写 TS interface 文件,与 pkg/models 一一对应
├── apis/
│   ├── axios.ts               # axios 实例 + 拦截器(envelope 解封装 + 401 跳转 + lang header)
│   ├── auth.ts                # login / logout / me
│   ├── me.ts                  # changePassword / setTushareToken / API tokens 增删
│   ├── portfolio.ts           # 列表 / 增删改备注
│   ├── stocks.ts              # 搜索 / 详情 / 日线
│   ├── analysis.ts            # 价差分析
│   ├── draft.ts               # 草稿 CRUD
│   └── admin.ts               # 用户管理 + 同步任务
├── stores/
│   ├── auth.ts                # User 状态 + login/logout/fetchMe
│   └── lang.ts                # 当前 lang(参考 mtkd/stores/lang.js)
├── intl/
│   ├── index.ts               # vue-i18n 实例
│   ├── lang.ts                # Langs / ElementLangs
│   └── langs/
│       ├── zh/index.ts        # 完整中文文案
│       └── en/index.ts        # 仅留 key + 空 value(后续补)
├── utils/
│   ├── message.ts             # wMessage 包装(去重 3s 内重复)
│   ├── storage.ts             # localStorage/sessionStorage 包装
│   └── format.ts              # 价格 / 涨跌幅格式化、CJK 数字补零
├── components/
│   ├── GIcon.vue              # 用 Element Plus icon 名包装
│   ├── GEllipsis.vue          # 文本溢出省略
│   ├── ConsoleMenu.vue        # 左上 logo + 中间股票菜单 + 底部 lang 切换
│   ├── UserMenu.vue           # 左下用户下拉(头像/名称 → 个人中心/退出)
│   ├── StockBasicCard.vue     # 顶部股票信息卡
│   ├── DailyBarTable.vue      # 日线表格(涨红跌绿)
│   ├── SpreadHistogram.vue    # 价差分布(Element progress 条 + 百分比)
│   ├── SpreadModelTable.vue   # 多窗口模型表
│   ├── TradePlanTable.vue     # 交易计划/反推预测表
│   └── DraftFormBlock.vue     # 今日 open/high/low/close 输入区
├── router/
│   └── index.ts
└── views/
    ├── LoginView.vue
    ├── StockListView.vue      # 替代当前 PortfolioView
    ├── stock/
    │   ├── StockDetailView.vue   # 外壳 + el-tabs
    │   ├── BasicTab.vue          # Tab 1
    │   └── StatisticsTab.vue     # Tab 2
    ├── profile/
    │   ├── ProfileView.vue       # 外壳 + 左侧子菜单
    │   ├── ProfileInfo.vue       # 只读:用户名、角色、注册时间
    │   ├── ChangePassword.vue
    │   ├── TushareToken.vue
    │   └── ApiTokens.vue
    ├── admin/
    │   ├── UsersView.vue
    │   └── SyncView.vue
    └── NotFound.vue
```

新增依赖:`vue-i18n@9`、`sass`(dev)。其它沿用已装的 element-plus / pinia / vue-router / axios / @element-plus/icons-vue。

### 3.2 整体布局

```
┌──────┬─────────────────────────────────────┐
│ logo │                                     │
│      │                                     │
│ 📈   │                                     │
│ 股票 │           main content              │
│      │       (router-view)                 │
│      │                                     │
│ 👤   │                                     │
│ 用户 │                                     │
│ 🌐 中 │                                     │
└──────┴─────────────────────────────────────┘
  65px
```

未登录(LoginView)不渲染侧栏,卡片居中。

### 3.3 配色与 SCSS 变量

`assets/css/index.scss` 头部 `:root`:

```scss
:root {
  --color-up: #f56c6c;     /* 涨红 */
  --color-down: #67c23a;   /* 跌绿 */
  --color-flat: #909399;   /* 平 */
  --sider-bg: #20242b;
  --sider-bg-hover: #313741;
  --sider-text: #acb3bf;
  --sider-active: #ffd04b;
  --content-bg: #f5f7fa;
  --card-radius: 4px;
}
```

主色保持 Element 默认 `#409EFF`,不覆盖 `--el-color-primary`。

### 3.4 路由设计

```ts
const routes = [
  { path: '/login', component: LoginView },
  { path: '/', redirect: '/stocks' },
  { path: '/stocks', component: StockListView, meta: { requiresAuth: true } },
  {
    path: '/stocks/:tsCode',
    component: StockDetailView,
    meta: { requiresAuth: true },
    children: [
      { path: '', name: 'StockBasic', component: BasicTab },
      { path: 'statistics', name: 'StockStatistics', component: StatisticsTab },
    ],
  },
  {
    path: '/profile',
    component: ProfileView,
    meta: { requiresAuth: true },
    children: [
      { path: '', component: ProfileInfo },
      { path: 'password', component: ChangePassword },
      { path: 'token', component: TushareToken },
      { path: 'api-tokens', component: ApiTokens },
    ],
  },
  { path: '/admin/users', component: UsersView, meta: { requiresAuth: true, requiresAdmin: true } },
  { path: '/admin/sync', component: SyncView, meta: { requiresAuth: true, requiresAdmin: true } },
  { path: '/:pathMatch(.*)*', component: NotFound },
]
```

`router.beforeEach` 保留现有 `requiresAuth` / `requiresAdmin` 守卫。

### 3.5 详情页 2 Tab(文档对齐)

**Tab 1 - 基础与价差** (`BasicTab.vue`)
- `StockBasicCard`:股票名/代码/行业/上市日期/最近收盘价(涨跌色)
- `SpreadHistogram`:6 种价差(高-开/开-低/高-低/开-收/高-收/低-收)的最近 N 日分布
- `DailyBarTable`:近 30 日 OHLCV + 6 价差列

**Tab 2 - 详细统计** (`StatisticsTab.vue`)
- `DraftFormBlock`(顶部):今日 open/high/low/close 输入 + 保存草稿 / 应用
- `SpreadModelTable`:4 窗口 × 6 价差均值表(综合均值行)
- `TradePlanTable`:历史/3月/1月/2周参考价 + 反推 + 综合均值列

现有 `AnalysisPanel.vue` / `HistoryPanel.vue` / `DraftPanel.vue` 删除,功能拆到上述子组件。

### 3.6 全局组件 `g-*`

`main.ts` 中注册(同 mtkd):
- `GIcon`:封装 Element Plus icon 字符串名 → 组件;`<g-icon name="TrendCharts" />`。
- `GEllipsis`:`<g-ellipsis :line="2">{{ longText }}</g-ellipsis>`。

不引入外部 iconfont 库,统一用 `@element-plus/icons-vue`。

### 3.7 国际化粒度

- 所有 view 标题、按钮文本、表头通过 `$t()`。
- 错误消息从后端 envelope.message 直接取(后端 `utils/msg.go` 根据 lang header 已返回对应语言)。
- zh-CN 文案完整;en-US `{}` 占位,i18n 配置 `fallbackLocale: 'zh'` 自动回退。
- 切换语言写 localStorage,axios 请求头带 `lang: zh` / `lang: en`。

---

## 4. 数据流与错误处理

### 4.1 调用栈

```
View (*.vue)
   ↓ 显式 import
apis/*.ts        // 函数签名 = TS interface
   ↓ axios.<method>
apis/axios.ts    // 实例 + 拦截器
   ↓ HTTP
/api/* (Gin handler)
   ↓ Bind→models.*Req
services/*       // 业务逻辑
   ↓
db (GORM models.*)
```

类型契约:
```
pkg/models/foo.go::Foo  ≡  web/src/types/api.ts::Foo
       ↑                          ↑
   CLI / handler 共享          views / apis / stores 共享
```

### 4.2 apis 层签名约定

```ts
// apis/portfolio.ts
import type { Portfolio, PortfolioReq } from '@/types/api'
import { $http } from './axios'

export const listPortfolio = (): Promise<Portfolio[]> => $http.get('/portfolio')
export const addPortfolio  = (req: PortfolioReq): Promise<void> => $http.post('/portfolio', req)
export const removePortfolio = (tsCode: string): Promise<void> => $http.delete(`/portfolio/${tsCode}`)
export const updatePortfolioNote = (tsCode: string, req: PortfolioReq): Promise<void> =>
  $http.put(`/portfolio/${tsCode}`, req)
```

View 层禁止直接 `$http.get('/x')`,必须经 `apis/*`,避免拼写漂移。

### 4.3 axios 拦截器

```ts
$http.interceptors.request.use(cfg => {
  cfg.headers.lang = langStore.lang  // 'zh' | 'en'
  return cfg
})

$http.interceptors.response.use(
  res => {
    // envelope: { requestID, code, message, data }
    if (res.data?.code === 200) return res.data.data
    wMessage.error(res.data?.message || 'unknown error')
    return Promise.reject(new Error(res.data?.message))
  },
  err => {
    if (err.response?.status === 401) {
      useAuthStore().logout()
      router.push('/login')
    } else {
      wMessage.error(err.message || '网络错误')
    }
    return Promise.reject(err)
  }
)
```

注意:与当前 `client.ts` 不同,响应拦截器直接返回 `res.data.data`,view 调用变为 `const list = await listPortfolio()`(单一返回值,无解构)。

### 4.4 错误处理边界

| 边界 | 处理 |
|---|---|
| 后端业务错(code !== 200) | 拦截器 `wMessage.error(message)` + reject;view 可 try/catch 做局部 UI |
| 401 未登录 | 拦截器自动 logout + 跳 `/login` |
| 网络异常 | 拦截器 `wMessage.error('网络错误')` + reject |
| 表单校验 | `el-form` rules,提交前同步校验 |
| Tushare token 缺失 | 后端返回特定 code,前端 `apis/me.ts` 引导到 `/profile/token` |

---

## 5. 测试

### 5.1 后端 / CLI

- **`pkg/models/models_test.go`** 增 `TestModelJSONFields`:table-driven,每个 struct marshal 后字段名集合 == 期望(`User` 必须含 `username`,不能含 `userName`)。
- **`pkg/stockd/http/*_test.go`** 已有,更新断言:用 `models.LoginReq` 等构造请求体,断言响应 JSON 字段名。
- **`pkg/cli/client/client_test.go`** 已有,补 portfolio GET 字段对齐用例(返回含 `userId`/`addedAt`,解析后 ID/AddedAt 非零)。

### 5.2 前端

- **vitest 单元测试**:
  - `apis/*.spec.ts`:用 `axios-mock-adapter` 断言请求/返回字段名。
  - `utils/format.spec.ts`:价格 / 涨跌幅格式化。
  - `components/*.spec.ts`:`DailyBarTable` 涨跌色 class 切换。
- **playwright e2e**(`web/e2e/`):
  - `login.spec.ts`:登录 → 跳 `/stocks`。
  - `stocks.spec.ts`:搜索 + 添加 + 进入详情。
  - `detail.spec.ts`:两个 Tab 切换 + 草稿输入 + 模型表渲染。
  - `profile.spec.ts`:改密码 + 设置 Tushare token + 创建/撤销 API token。

### 5.3 契约同步检查

- 后端字段集合靠 §5.1 的 `TestModelJSONFields` 锁死。
- `web/src/types/api.ts` 顶部注释 `// keep in sync with pkg/models`;每个 interface 上注释 `@see pkg/models/foo.go::Foo`。
- 后续(本 spec 不实施)可写 `scripts/check-api-sync.sh`:grep 提取两边字段做 diff,CI 跑。

---

## 6. 风险与回滚

| 风险 | 严重 | 缓解 |
|---|---|---|
| 字段重命名破坏 session 缓存的 `user.userName` | 中 | 后端 JWT/session 不依赖字段名;`stores/auth` 加 schema 版本 key,版本不匹配清空 localStorage |
| 一次性重写期间无法独立发版 | 中 | feature branch + e2e 全绿 + 人工 smoke 后合 master |
| en-US 留空导致英文模式空白 | 低 | vue-i18n `fallbackLocale: 'zh'` 自动回退 |
| CLI portfolio 输出格式变化 | 低 | 输出仍为 `<tsCode>\t<note>`,只是内部解析类型修正 |
| 删除旧 view 丢功能 | 中 | 严格按 §7 旧→新映射执行;e2e 兜底 |

---

## 7. 旧 → 新功能映射

| 旧组件 / 视图 | 新位置 |
|---|---|
| `components/AnalysisPanel.vue` | `views/stock/StatisticsTab.vue` + `SpreadModelTable` + `TradePlanTable` |
| `components/HistoryPanel.vue` | `views/stock/BasicTab.vue` + `DailyBarTable` |
| `components/DraftPanel.vue` | `views/stock/StatisticsTab.vue` 顶部 `DraftFormBlock` |
| `views/PortfolioView.vue` | `views/StockListView.vue` |
| `views/SettingsView.vue` | `views/profile/`(拆 4 子页) |
| `views/admin/UsersView.vue` | `views/admin/UsersView.vue`(保留路径,套新样式) |
| `views/admin/SyncView.vue` | `views/admin/SyncView.vue`(同上) |
| `api/client.ts` | `apis/axios.ts` + `apis/*.ts` 拆分 |
| 内联类型(`AnalysisPanel` 用 `any`、`stores/portfolio` 用 snake_case) | `types/api.ts` 统一 |

---

## 8. 完成定义(DoD)

- [ ] `pkg/models/*` 所有 inline DTO 收敛(`auth.go` / `me.go` 新增完毕)
- [ ] `pkg/models` JSON 字段全部标准 camelCase(`username` / `userId` / `stockId`)
- [ ] `pkg/models/doc.go` 包注释提示前后端契约对齐
- [ ] `pkg/stockd/http/*.go` 不再有 inline Req/Resp struct
- [ ] `pkg/cli/cmd/*.go` 类型与 handler 对齐,不再 inline(包括 `cli/cmd/portfolio.go` 改用 `Portfolio` 类型)
- [ ] `pkg/models/models_test.go::TestModelJSONFields` 通过
- [ ] `go test -race ./...` 全绿
- [ ] `web/src/types/api.ts` 字段名 100% 匹配 `pkg/models`
- [ ] 前端所有 view 经 `apis/*.ts` 调用,不直接用 axios
- [ ] 涨红跌绿色变量在所有数值展示处生效
- [ ] `web/src/intl/` 完整 zh-CN 文案,en-US 占位
- [ ] vitest 单元测试全绿
- [ ] playwright e2e(login / stocks / detail / profile)全绿
- [ ] 人工 smoke:登录 → 列表 → 详情两 Tab → 个人中心改密码 → 切换语言 → 登出

---

## 9. 后续阶段(本 spec 不实施)

- en-US 完整翻译填充
- echarts 接入,价差分布图升级为柱状图
- 深色主题切换
- 移动端响应式
- `scripts/check-api-sync.sh` CI 检查
