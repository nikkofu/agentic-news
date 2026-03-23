# The Digital Sanctuary - Golang API 设计与接口规范

本文档定义了前端 SPA 页面与后端 Golang API 交互的契约。所有的 RESTful API 请以 `/api/v1` 作为基础路径。数据传输严格采用 JSON 格式。

## 核心数据结构 (Golang ORM 模型参考)

```go
package models

import "time"

// User 表示系统核心用户体系
type User struct {
    ID             string    `json:"id" gorm:"primaryKey"`
    Email          string    `json:"email" gorm:"uniqueIndex"`
    Tier           string    `json:"tier"` // Free, Scholar, Patron
    FocusGoalMins  int       `json:"focus_goal_mins"`
    CreatedAt      time.Time `json:"created_at"`
}

// UserDomain 表示用户在 Onboarding 阶段关注的知识领域
type UserDomain struct {
    ID       string `json:"id" gorm:"primaryKey"`
    UserID   string `json:"user_id" gorm:"index"`
    Domain   string `json:"domain"` // 如 Artificial Intelligence
}

// Article 表示资讯或文章内容
type Article struct {
    ID          string    `json:"id" gorm:"primaryKey"`
    Title       string    `json:"title"`
    Summary     string    `json:"summary"`
    Content     string    `json:"content"`
    PublishTime time.Time `json:"publish_time"`
    Score       float64   `json:"score"` // 推荐算法计算的权重分数
}

// GraphNode 与 GraphEdge 表示 Vault 星图谱系数据
type GraphNode struct {
    ID    string `json:"id"`
    Type  string `json:"type"` // 取值: Concept, Article, Highlight
    Label string `json:"label"`
}

type GraphEdge struct {
    Source string `json:"source"`
    Target string `json:"target"`
    Weight int    `json:"weight"`
}
```

---

## 接口路由表及入参出参说明

### 1. 认证模块 (Authentication)
*支撑 `#/login` 的 Magic Link 逻辑*

**1.1 请求魔法链接**
- **Endpoint**: `POST /auth/magic-link`
- **Request**: `{ "email": "user@example.com" }`
- **Response**: `200 OK` (返回状态即可，勿返回凭证内容)

**1.2 校验令牌并签发 JWT**
- **Endpoint**: `POST /auth/verify`
- **Request**: `{ "token": "abc123xyz" }`
- **Response**:
  ```json
  { 
      "access_token": "jwt_string", 
      "refresh_token": "jwt_string", 
      "user": { "id": "uuid", "email": "user@example.com", "tier": "Free" } 
  }
  ```

### 2. 用户偏好与入驻 (User Onboarding & Settings)
*支撑 `#/onboarding` 与 `#/settings`*

**2.1 获取/更新用户个人设置**
- **Endpoint**: `GET /users/me/preferences` | `PUT /users/me/preferences`
- **PUT Payload**:
  ```json
  {
      "domains": ["Artificial Intelligence", "Stoic Philosophy", "Systems Thinking"],
      "focus_goal_mins": 45,
      "sources": {
          "rss": ["https://tech.com/rss"],
          "twitter_handle": "@scholar"
      }
  }
  ```

### 3. 早报与阅读核心流 (Daily Briefing & Feed)
*支撑 `#/briefing` 与 `#/category`*

**3.1 获取每日速递 (Daily Briefing 动态渲染数据)**
- **Endpoint**: `GET /feed/daily`
- **Response**:
  ```json
  {
      "top_insight": {
          "title": "The Architecture of Silence",
          "summary": "Explore how physical environments dictate...",
          "image_url": "https://..."
      },
      "butler_suggestion": {
          "endurance_pct": 82,
          "progress_mins": 45,
          "target_mins": 60
      },
      "curated_articles": [ { "id": "1", "title": "The Moral Ghost..." } ],
      "quick_reads": [ { "id": "2", "title": "Digital Stoicism", "read_time": 6 } ]
  }
  ```

**3.2 获取 AI 文章沉浸式机器综述 (AI Synthesis)**
- **Endpoint**: `GET /articles/:id/synthesis`
- **说明**: 该接口对于长文章的 AI 分析较慢，**强烈建议后端采用 Server-Sent Events (SSE)** 流式下发 `thought_chunk` 文本给前台以呈现打字机效果。全量结构：
  ```json
  {
      "core_thesis": "...",
      "key_nodes": ["Kant", "Alignment"],
      "contradictions": ["Human Values vs Code"]
  }
  ```

### 4. 知识星图数据下发 (Vault Graph)
*支撑 `#/vault`，需处理大量的节点计算*

**4.1 抓取个人专属的数字知识网络图表**
- **Endpoint**: `GET /vault/graph`
- **QueryParams**: `?filter_domain=ai&timeframe=30d`
- **Response**: 节点数可能极大，Go 后端需做好缓存并精简 JSON 体积。
  ```json
  {
      "nodes": [ { "id": "n1", "type": "Concept", "label": "LLM Alignment", "val": 20 } ],
      "edges": [ { "source": "n1", "target": "n2", "weight": 5 } ]
  }
  ```

### 5. 注意力分析仪表盘 (Analytics)
*支撑 `#/analytics` 本地时间序列数据的聚合*

**5.1 获取多维度专注力统计分析**
- **Endpoint**: `GET /analytics/summary`
- **Response**:
  ```json
  {
      "cognitive_load_index": 7.4,
      "peak_focus_hours": [8, 9, 10], 
      "spectrum_analysis": {
          "deep_work": 65,
          "exploration": 20,
          "distraction": 15
      },
      "echo_chamber_alert": true
  }
  ```

### 6. 生成数字文物快照 (Share Artifact)
*支撑 `#/share` 卡片裂变生成引擎*

**6.1 由后端生成/保存知识卡片快照防篡改**
- **Endpoint**: `POST /artifacts`
- **Request**:
  ```json
  {
      "source_text": "Silence is the sleep that nourishes wisdom.",
      "author": "Francis Bacon",
      "style_preset": "parchment"
  }
  ```
- **Response**:
  ```json
  { "artifact_id": "8b9cad0e", "artifact_url": "https://cdn.sanctuary/...png", "shareable_link": "https://host/artifacts/8b9cad0e" }
  ```

### 7. 订阅收银台与账单 (Treasury Checkout & Billing)
*支撑 `#/checkout` 商用闭环的核心交易环节*

**7.1 创建支付意向 (调用第三方支付 SDK, 如 Stripe Intent)**
- **Endpoint**: `POST /billing/intent`
- **Request**: `{ "tier_id": "scholar_monthly" }`
- **Response**:
  ```json
  { "client_secret": "pi_1Mxxxx_secret_xxxxx", "amount": 1900 }
  ```

**7.2 获取账单历史**
- **Endpoint**: `GET /billing/history`
- **Response**:
  ```json
  [ { "invoice_id": "TRSY-8821", "amount": 19.00, "status": "paid", "date": "2024-09-12T00:00:00Z" } ]
  ```

### 8. 系统的神谕消息 (Oracle Notifications)
*支撑 `#/notifications` 消息侧抽屉*

**8.1 分页抓取通知墙列表**
- **Endpoint**: `GET /notifications`
- **Response**:
  ```json
  {
      "notifications": [
          { "id": "notif_1", "type": "broadcast", "title": "Upgrade 4.0", "read": false, "created_at": "..." },
          { "id": "notif_2", "type": "butler_intel", "title": "Cognitive Pattern", "read": false }
      ]
  }
  ```
**8.2 批量全域标为已阅**
- **Endpoint**: `PUT /notifications/read-all`
- **Response**: `200 OK`
