# 生产环境发布与部署架构指南 (Deployment Architecture)

本指南针对前端静态资源 (Digital Sanctuary UI) 以及即将用 Golang 开发的强大后端 (API Services) 进行生产级别定义。

## 1. 静态前端交付网 (CDN Edge Deployment)
鉴于本项目前端是一个纯粹的基于 HTML/CSS/JS 的 SPA。
*   **承载方式**: 绝对**不**建议用 Golang 来托管这些静态文件（如 `http.FileServer`），这会浪费宝贵的连接池资源。
*   **最佳实践**: 将整个 `/apps` 文件夹上传至 Amazon S3 / Cloudflare Pages / Vercel 等平台，其天然具备边缘节点加速且成本极低。
*   **路由兜底设置**: 因采用的是 Hash 模式路由 (`/#/home`)，天然兼容所有静态服务器，无需额外配置 `rewrite /index.html` 操作。但如果未来改为 History Router，必须要强制 404 跳转根目录。

## 2. Golang API 容器化布署 (Dockerizing Go)
*   为了平滑拓展，后端 API 会被打包为极简的 Alpine 或 Scratch Docker Image。
*   Dockerfile 参考：
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```
*   **集群建议**: 建议部署在 Kubernetes (K8S) 或 AWS ECS，且前置使用 Nginx Ingress / ALB 进行负载均衡及 TLS (HTTPS) 终结卸载。

## 3. 跨域防护 (CORS Protocol)
当前端处于 `app.digital-sanctuary.com` 后端挂在 `api.digital-sanctuary.com` 时：
1. Golang 全局中间件需严格允许跨域访问来源 (Allow-Origins)。
2. 因为采用 `HTTPOnly` 的 Session Cookie (刷新 Token)，因此 `Access-Control-Allow-Credentials` 须设为 `true`。
3. 暴露 Authorization 首部。
