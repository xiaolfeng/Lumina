package handler

import (
	"encoding/base64"
	"fmt"
	"html"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const navigateShim = `<script data-lumina-nav="1">
(function(){
  document.addEventListener("click", function(e){
    if (e.defaultPrevented) return;
    var a = e.target && e.target.closest ? e.target.closest("a") : null;
    if (!a) return;
    if (a.hasAttribute("download")) return;
    var target = a.getAttribute("target");
    if (target && target !== "_self") return;
    var href = a.getAttribute("href");
    if (!href || href.charAt(0) === "#" || /^(https?:|mailto:|javascript:|tel:)/i.test(href)) return;
    if (!/\.(html|htm|md|tsx|jsx)([?#].*)?$/i.test(href.split("/").pop() || href)) return;
    try { parent.postMessage({ type: "lumina:navigate", href: href }, "*"); } catch (err) {}
  }, false);
})();
</script>`

// IsDocumentRequest 判断是否为顶层文档导航。
// iframe / style / script 等资源请求必须直出文件，不能回落 SPA。
// 前端预览 iframe 会追加 lumina_frame 查询参数：带参请求一律按非顶层文档处理，
// 避免老浏览器缺少 Sec-Fetch-Dest 时被 Accept 误判为顶层文档而回落 SPA。
// 缺少 Sec-Fetch-Dest 时仅当 Accept 含 text/html 才视为文档，避免 curl 拉 CSS 时误返回 index.html。
func IsDocumentRequest(ctx *gin.Context) bool {
	if ctx.Query("lumina_frame") != "" {
		return false
	}
	dest := strings.ToLower(strings.TrimSpace(ctx.GetHeader("Sec-Fetch-Dest")))
	switch dest {
	case "document":
		return true
	case "":
		accept := strings.ToLower(ctx.GetHeader("Accept"))
		return strings.Contains(accept, "text/html")
	default:
		return false
	}
}

func isTSXMime(mimeType string) bool {
	lower := strings.ToLower(mimeType)
	return strings.Contains(lower, "typescript-jsx") || strings.Contains(lower, "text/jsx")
}

func writeServedFile(ctx *gin.Context, filename, mimeType, content string) {
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
	ctx.Header("Access-Control-Allow-Headers", "*")
	mime := strings.ToLower(mimeType)
	body := content

	// 若请求的是 TSX/JSX 且非纯文本 raw 模式，自动包裹为交互式虚拟宿主 HTML
	// 若客户端明确传了 raw 参数或 Sec-Fetch-Dest 为 empty (fetch 请求)，直出原始纯文本
	isRawFetch := ctx.Query("raw") != "" || strings.EqualFold(ctx.GetHeader("Sec-Fetch-Dest"), "empty")
	if isTSXMime(mime) && !isRawFetch {
		mime = "text/html; charset=utf-8"
		body = renderTSXVirtualHost(filename, content)
	}

	if strings.HasPrefix(mime, "text/html") || mime == "image/svg+xml" || strings.Contains(mime, "javascript") {
		ctx.Header("Content-Security-Policy", "sandbox allow-scripts")
	}
	if strings.HasPrefix(mime, "text/html") {
		body = injectNavigateShim(body)
	}
	ctx.Data(http.StatusOK, mime, []byte(body))
}

func renderTSXVirtualHost(filename, content string) string {
	escapedTitle := html.EscapeString(filename)
	encodedSource := base64.StdEncoding.EncodeToString([]byte(content))
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>%s · Lumina Preview</title>
  <link rel="stylesheet" href="style.css">
  <script src="https://cdn.tailwindcss.com"></script>
  <script crossorigin src="https://unpkg.com/react@18.3.1/umd/react.production.min.js"></script>
  <script crossorigin src="https://unpkg.com/react-dom@18.3.1/umd/react-dom.production.min.js"></script>
  <script src="https://unpkg.com/@babel/standalone@7.26.9/babel.min.js"></script>
  <style>
    html, body, #root { min-height: 100%%; margin: 0; }
  </style>
</head>
<body class="bg-slate-50 text-slate-900 font-sans antialiased">
  <div id="root"></div>
  <script id="__lumina_tsx_source__" type="text/plain" data-encoding="base64">%s</script>
  <script>
    (function() {
      var rootEl = document.getElementById('root');
      try {
        if (typeof window.React === 'undefined' || typeof window.ReactDOM === 'undefined') {
          throw new Error('React 或 ReactDOM 运行时加载失败，请检查网络连接或 CDN 状态');
        }
        var React = window.React;
        var ReactDOM = window.ReactDOM;

        var sourceEl = document.getElementById('__lumina_tsx_source__');
        var rawBase64 = (sourceEl.textContent || '').trim();
        var code = '';
        try {
          code = decodeURIComponent(escape(atob(rawBase64)));
        } catch (decodeErr) {
          code = atob(rawBase64);
        }

        var transformed = Babel.transform(code, {
          presets: [
            ['env', { modules: 'commonjs' }],
            ['react', { runtime: 'classic' }],
            ['typescript', { isTSX: true, allExtensions: true }]
          ],
          filename: 'component.tsx'
        }).code;

        var exports = {};
        var module = { exports: exports };
        function require(name) {
          if (name === 'react') return React;
          if (name === 'react-dom' || name === 'react-dom/client') return ReactDOM;
          return {};
        }

        var runner = new Function(
          'React', 'ReactDOM', 'module', 'exports', 'require',
          'useState', 'useEffect', 'useRef', 'useMemo', 'useCallback', 'useContext', 'useReducer',
          transformed
        );
        runner(
          React, ReactDOM, module, exports, require,
          React.useState, React.useEffect, React.useRef, React.useMemo, React.useCallback, React.useContext, React.useReducer
        );

        var Component = module.exports.default || module.exports.App || module.exports.Main;
        if (!Component) {
          for (var key in module.exports) {
            if (typeof module.exports[key] === 'function' && key.charCodeAt(0) >= 65 && key.charCodeAt(0) <= 90) {
              Component = module.exports[key];
              break;
            }
          }
        }

        if (Component) {
          var root = ReactDOM.createRoot(rootEl);
          root.render(React.createElement(Component));
        } else {
          rootEl.innerHTML = '<div style="padding:2rem;font-family:sans-serif;color:#e11d48;background:#fff1f2;border:1px solid #ffe4e6;border-radius:8px;margin:1.5rem;">' +
            '<h3 style="margin:0 0 0.5rem 0;font-size:1.125rem;font-weight:600;">未找到可渲染的 React 组件</h3>' +
            '<p style="margin:0;font-size:0.875rem;color:#4c0519;">请确保文件中包含 <code>export default function App() { ... }</code> 或大写命名的组件函数。</p>' +
            '</div>';
        }
      } catch (err) {
        console.error('[Lumina TSX Runtime Error]', err);
        rootEl.innerHTML = '<div style="padding:2rem;font-family:monospace;color:#991b1b;background:#fef2f2;border:1px solid #fecaca;border-radius:8px;margin:1.5rem;">' +
          '<h3 style="margin:0 0 0.5rem 0;font-size:1.125rem;font-weight:600;">TSX 编译或渲染异常</h3>' +
          '<pre style="margin:0;font-size:0.8125rem;white-space:pre-wrap;word-break:break-all;">' + (err.stack || err.message) + '</pre>' +
          '</div>';
      }
    })();
  </script>
</body>
</html>`, escapedTitle, encodedSource)
}

func injectNavigateShim(html string) string {
	lower := strings.ToLower(html)
	if idx := strings.LastIndex(lower, "</body>"); idx >= 0 {
		return html[:idx] + navigateShim + html[idx:]
	}
	return html + navigateShim
}
