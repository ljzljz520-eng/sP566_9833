package web

import (
	"net/http"

	"pharmacycounter/model"
)

const IndexHTML = `<!doctype html>
<html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>药房取药台</title><link rel="stylesheet" href="/style.css"></head>
<body><main><header><h1>药房取药台</h1><p>待取药 · 已叫号 · 已完成</p></header><section class="lists"><article><h2>待取药</h2><ul id="pending"></ul></article><article><h2>已叫号</h2><ul id="called"></ul></article><article><h2>已完成</h2><ul id="completed"></ul></article></section></main><script src="/app.js"></script></body></html>`

const StyleCSS = `:root{font-family:system-ui,sans-serif;color:#18212b;background:#f4f6f8}body{margin:0}main{max-width:1100px;margin:0 auto;padding:32px}header{border-bottom:1px solid #d5dbe1;margin-bottom:24px}.lists{display:grid;grid-template-columns:repeat(3,1fr);gap:16px}article{background:#fff;border:1px solid #d5dbe1;padding:16px;min-height:260px}h1{margin:0 0 8px}h2{font-size:1rem}li{list-style:none;padding:8px 0;border-bottom:1px solid #edf0f2}@media(max-width:700px){.lists{grid-template-columns:1fr}main{padding:16px}}`

const AppJS = `const labels={pending:"pending",called:"called",completed:"completed"}; for(const id of Object.keys(labels)){document.getElementById(id).innerHTML='<li>等待服务台数据</li>'}`

func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/health", HealthHandler())
	mux.Handle("/api/tickets", TicketHandler(func() []model.PharmacyTicket { return []model.PharmacyTicket{} }))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(IndexHTML))
	})
	mux.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		_, _ = w.Write([]byte(StyleCSS))
	})
	mux.HandleFunc("/app.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
		_, _ = w.Write([]byte(AppJS))
	})
	return mux
}
