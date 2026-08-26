# BENZHI_README

基于 Go 实现的城市树木病虫害样本放行工作台 Web 项目，一款后端服务，把树木病虫害样本从现场采集推进到鉴定、整改、复核和处置放行。

## 项目说明
- 项目：benzhi-project-d47336ac-7d45-4f6c-8f37-a3bbd8e1f71c
- 项目用途：把树木病虫害样本从现场采集推进到鉴定、整改、复核和处置放行。
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
