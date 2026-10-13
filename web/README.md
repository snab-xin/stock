# ashare-web

A 股数据质检看板：Vite + React + TypeScript + shadcn 风格组件 + ECharts。

## 开发

先启动 API，再启动前端（两个终端）：

```powershell
cd d:\workspace\stock\ashare
go run ./cmd/ashare web
```

```powershell
cd d:\workspace\stock\ashare\web
npm install
npm run dev
```

浏览器打开 http://127.0.0.1:5173 。Vite 会把 `/api` 代理到 `127.0.0.1:8080`。

## 页面

- `/` 行情质检：搜索证券、前复权/未复权 K 线、涨跌停/停牌标记、缺日
- `/ingest` 采集任务：只读 `ingest_run`

灌数仍用 `ashare ingest`，本前端不负责下单或调 Tushare。
