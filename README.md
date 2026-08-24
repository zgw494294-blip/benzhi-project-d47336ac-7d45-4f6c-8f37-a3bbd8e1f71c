# 城市树木病虫害样本放行工作台

本项目为城市绿化团队提供浏览器工作台，将树木病虫害样本从现场采集推进到完整性检查、专家鉴定、整改、独立复核和处置放行。所有状态变化写入带 SHA-256 连续摘要的 JSON Lines 审计账本，放行后生成可在线校验的冻结凭据。

## 构建、运行与测试

```bash
go test ./...
go run . -selfcheck
go run . -addr=127.0.0.1:19081
```

浏览器访问 `http://127.0.0.1:19081/workbench`。也可以通过 `PORT` 环境变量指定端口，例如 `PORT=19082 go run .`。数据默认保存在 `.refill-data`，可使用 `-data` 指定目录。

## 业务接口

- `GET /api/batches`：批次目录。支持 `status`、`q`、`role`、`page` 和 `pageSize` 组合筛选；按更新时间倒序返回 `total`、`page`、`pageSize`、`hasNext` 和 `batches`。角色可使用 `collector`、`expert`、`reviewer`、`admin` 或对应中文名称。
- `POST /api/batches`：建立样本批次。
- `POST /api/batches/{id}/evidence`：提交现场证据。证据可先进入待检查状态，检查流程会逐条给出完整性分数和警告。
- `POST /api/batches/{id}/screening`：执行完整性与环境阈值检查，请求必须包含鉴定员角色、`expectedVersion` 和 `idempotencyKey`。
- `POST /api/batches/{id}/review`：提交专家鉴定。
- `POST /api/batches/{id}/rectification`：提交整改闭环。
- `POST /api/batches/{id}/release`：冻结并签发凭据。`plan` 支持使用顿号、逗号或分号分隔步骤，可选传入 `planName` 和 `owner`；未传时分别沿用标准方案名称和复核人。
- `GET /api/certificates/{id}/verify`：重算冻结摘要、核对放行事件并校验审计事件链。
