# 整体架构设计规范 (System Architecture)

The Digital Sanctuary 采用彻底的前后端分离架构 (Decoupled Single Page Application Architecture)。针对这种架构形态，以下是供给给由 Golang 引领的后端梯队的战略规划建议：

## 1. 技术栈选型
*   **前端结构**: 基于原生 DOM 与 HTML5 History / Hash Router 路由，抛弃重型框架以保障“庇护所”的无感秒级加载。
*   **前端样式**: 移植自 Tailwind CSS 的 `sanctuary.css` 构建的 CSS Variables 系统。
*   **后端开发语言**: Golang 工具链 (推荐 1.21 以上引擎)，利用 Go 的 Goroutines 满足全动态高并发下海量用户的 RSS 和信息源抓取诉求。
*   **API 协议**: JSON RESTful APIs；部分 AI 总结反馈接口必须采用 **Server-Sent Events (SSE)**，便于将 LLM 生成数据的字节流推送到前端实现极其平滑的打字机交互（不推荐用 WebSocket 应对普通的单向下发订阅）。
*   **底层数据库选型**: 
    - PostgreSQL: 存储低频稳定关系型业务数据（如用户信息、账单支付履历、文档图文库实体）。
    - Redis: 专用于 JWT 过期黑名单维护、极高频次的 `attention analytics` 打点计数记录，以及外网 API 的 Rate Limiting 保障。
    - Graph DB (如 Neo4j，可选进阶): 用来完美替代关系型数据库以渲染 `The Vault` (知识星图) 里的深层级 Node-Edge 知识网络映射。

## 2. 鉴权与会话管理 (Authentication System)

前端的设计初衷在于取消了繁琐的密码登录密码找回环节。纯动态开发时请务必落实：
**Magic Link (魔法链接) + JWT 双令牌机制**
1. 初次准入：用户在 `#/login` 输入邮箱。
2. 内部扭转：Golang 后端基于加密时间戳生成有效期极短（如 15分钟）的校验验证码，通过 SMTP 发送包含了带有 `?token=xxx` 参数的验证邮件。
3. 校验颁布：访问被拦截提取 `token` 发给后端 `POST /api/v1/auth/verify`。
4. Go API 判断此有效后，颁发：短期令牌 `access_token` (放置于内存/Authorization头部用于各种 API 获取数据)，与长期令牌 `refresh_token` (设定为 HTTP-Only Cookie 提供无感持久续期)。

## 3. 计费与订阅架构闭环 (Billing Automation)

考虑到 `#/checkout` 存在严格的 Scholarship 和 Patron 的权限划分，不能只靠前台控制。
* 客户端发起：前端携带所选套餐，通过 `POST /billing/intent` 请求 Stripe 服务得到交易单凭证。
* 收银台：前端依托于生成的 `client_secret` 在隔离网关的沙盒中执行用户的扣款信息输入。
* 履约通知：**绝对不可依仗前端请求“通知”服务端已支付。** Golang API 后端必须暴露一条安全的、带有验签逻辑的 Webhook 服务端口（例如 `POST /api/v1/webhooks/stripe`）来捕捉 `invoice.payment_succeeded`。系统通过这一步骤来激活和写入 PostgreSQL 里的数据库特权身份标志。

## 4. 人工管家与AI处理后台 (AI Butler Worker Pool)

“The Oracle (神谕流)” 以及智能阅读总结都极大依赖了 AI（大语言模型调优）。
为了不使用户的主线访问在等待 AI 反应中 Time Out，处理策略应当为：
* 设立专线的 Go Worker Pool：依托 Asynq 获取系统内部产生的数据消费任务以进行长文本推断运算。
* 状态轮询/异步推送：对于较耗时的 AI 深加工知识网络关联挖掘，后台静默得出结论后，写入 Notification 表。当用户客户端下次查询 `/api/v1/notifications` 或通过长链轮询时，及时呈现神谕的提醒提示（Butler Intel）。
