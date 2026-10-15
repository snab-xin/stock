# ashare web

基于 [shadcn-admin](https://github.com/satnaing/shadcn-admin) 的 A 股数据看板前端。

## 开发

需先启动后端 API：

```powershell
go run ./cmd/ashare web
```

再启动前端：

```powershell
cd web
pnpm install
pnpm dev
```

打开 http://127.0.0.1:5173 。`/api` 会代理到 `http://127.0.0.1:8080`。

## 页面

- `/` 行情质检
- `/cb` 可转债列表（涨红跌绿）
- `/cb/$code` 可转债详情 / 十大持有人
- `/ingest` 采集任务
- `/backtest` 最小回测

## 构建

```powershell
pnpm build
```
