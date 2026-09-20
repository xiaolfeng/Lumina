package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const navigateShim = `<script data-lumina-nav="1">
(function(){
  document.addEventListener("click", function(e){
    var a = e.target && e.target.closest ? e.target.closest("a") : null;
    if (!a) return;
    var href = a.getAttribute("href");
    if (!href || href.charAt(0) === "#" || /^(https?:|mailto:|javascript:)/i.test(href)) return;
    if (!/\.(html|htm|md)([?#].*)?$/i.test(href.split("/").pop() || href)) return;
    e.preventDefault();
    try { parent.postMessage({ type: "lumina:navigate", href: href }, "*"); } catch (err) {}
  }, true);
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

func writeServedFile(ctx *gin.Context, mimeType, content string) {
	ctx.Header("Cache-Control", "no-cache")
	mime := strings.ToLower(mimeType)
	if strings.HasPrefix(mime, "text/html") || mime == "image/svg+xml" || strings.Contains(mime, "javascript") {
		ctx.Header("Content-Security-Policy", "sandbox allow-scripts")
	}
	body := content
	if strings.HasPrefix(mime, "text/html") {
		body = injectNavigateShim(content)
	}
	ctx.Data(http.StatusOK, mimeType, []byte(body))
}

func injectNavigateShim(html string) string {
	lower := strings.ToLower(html)
	if idx := strings.LastIndex(lower, "</body>"); idx >= 0 {
		return html[:idx] + navigateShim + html[idx:]
	}
	return html + navigateShim
}
