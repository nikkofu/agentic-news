# 前端数据流与状态管理规范 (State Management)

在我们当前的纯 Vanilla JS 架构下，避免引入 Redux 等复杂巨物，需采用轻量且健壮的数据层进行管理，使得 `app.js` 这个单页路由入口能够长久维护。

## 1. 全局状态容器 (`window.AppState`)
建议在 `app.js` 或专门的 `store.js` 内初始化一个全局不可变的数据集：
```javascript
const AppState = {
    user: null,               // 包含 tier, email 等。由 /auth/verify 给定数据后持久化
    accessToken: null,        // 存在内存中，关闭便丢失，用于 HTTP Barer Header
    notifications: [],        // 缓存的神谕推流数据
    selectedDomains: [],      // 临时选择的偏好
};
```

## 2. 跨页面的持久化通信 (Storage)
对于无需强安全性加密但又需要在页面刷新后依旧保持的配置（例如主题色偏好，引导流进入状态）：
- 采用 `localStorage`。
- **示例**：`localStorage.setItem('sanctuary_theme', 'dark')`

对于认证信息：
- `refresh_token` 应当由 Golang 后端在颁发时设为 `HTTPOnly: true, Secure: true` 的 Cookie。前端代码**不该**也**无法**直接接触它，只在 Fetch 请求过期时由浏览器自动带着这个 Cookie 访问 `/auth/refresh` 获取崭新的 `access_token`。

## 3. DOM 数据绑定策略
因为缺少 Vue/React 的响应式虚拟 DOM，建议使用 EventTarget 自发布订阅：
1. UI 元素触发操作 (如点击 Checkout)。
2. Fetch 调用 API (如拿取 Client Secret)。
3. 操作完更改 AppState，并发射一个 CustomEvent (如 `document.dispatchEvent(new CustomEvent('USER_UPDATED'))`)。
4. HTML `<aside>` 等区块内部包含的自调用函数或 MutationObserver 监听到此事件，并自动重绘界面部分数据。

通过这种“单向数据流”可彻底避免意大利面条级别的 DOM 查询和修改。
