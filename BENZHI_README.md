# BENZHI_README

基于 Go 实现的城市树木病虫害样本放行工作台 Web 项目，一款后端服务，城市树木病虫害样本放行工作台已完整实现从样本建档、现场证据采集、自动检查、专家鉴定、整改、独立复核到冻结放行凭据校验的唯一业务流程。服务默认监听 127.0.0.1:19081，支持 -addr 和 PORT 配置，浏览器入口为 /workbench，审计数据使用带连续 SHA-256 摘要的 JSON Lines 账本并通过原子快照恢复。

## 项目说明
- 项目：benzhi-project-d47336ac-7d45-4f6c-8f37-a3bbd8e1f71c
- 项目用途：城市树木病虫害样本放行工作台已完整实现从样本建档、现场证据采集、自动检查、专家鉴定、整改、独立复核到冻结放行凭据校验的唯一业务流程。服务默认监听 127.0.0.1:19081，支持 -addr 和 PORT 配置，浏览器入口为 /workbench，审计数据使用带连续 SHA-256 摘要的 JSON Lines 账本并通过原子快照恢复。
- Go 工具链：`golang:1.26`
- 前端工具链：原生 HTML、CSS 和 JavaScript，由 Go 服务直接提供

## 标准构建、运行和测试命令
进入容器后执行：
cd '/app' && GOTOOLCHAIN=local go build ./...
cd /app && GOTOOLCHAIN=local go run . -selfcheck
cd '/app' && GOTOOLCHAIN=local go test ./...

## Docker 构建和进入容器
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh benzhi-project-d47336ac-7d45-4f6c-8f37-a3bbd8e1f71c-amd64 linux/amd64
./build_benzhi_docker.sh benzhi-project-d47336ac-7d45-4f6c-8f37-a3bbd8e1f71c-arm64 linux/arm64
docker run -it benzhi-project-d47336ac-7d45-4f6c-8f37-a3bbd8e1f71c-amd64:latest

## 题目验证命令
1. 预期退出码 0：`go test ./...`
2. 预期退出码 0：`go run . -selfcheck`
