package logic

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/repository"
)

// resolveRuntimeDomain 解析面向用户的站点访问域名。
//
// 优先使用 Info 表中的 site.domain；未配置时回退到本地监听地址。该方法只负责
// 解析站点根地址，具体业务页面路径由各模块自行拼装。
func resolveRuntimeDomain(ctx context.Context, infoRepo *repository.InfoRepo) string {
	if domain, xErr := infoRepo.GetByKey(ctx, bConst.InfoKeySiteDomain); xErr == nil && domain != "" {
		return strings.TrimRight(domain, "/")
	}

	host := xEnv.GetEnvString(xEnv.Host, "localhost")
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "localhost"
	}
	port := xEnv.GetEnvString(xEnv.Port, "8080")
	return fmt.Sprintf("http://%s:%s", host, port)
}

// buildPreviewURL 构造 Preview 会话访问地址；filename 非空时生成文件深链。
func buildPreviewURL(domain, hash, filename string) string {
	query := url.Values{}
	query.Set("session", hash)
	if filename != "" {
		query.Set("file", filename)
	}
	return fmt.Sprintf("%s/preview?%s", strings.TrimRight(domain, "/"), query.Encode())
}
