package prepare

import (
	"fmt"
	"strconv"
	"time"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
)

// preparePreviewExpiresAt 幂等回填历史预览会话的过期时间
//
// 没有 expires_at 的行写成 now + 当前 preview.session.ttl（读失败用 7 天）。
func (p *Prepare) preparePreviewExpiresAt() {
	ttlSec := 604800
	var info entity.Info
	if err := p.db.WithContext(p.ctx).Where("key = ?", bConst.InfoKeyPreviewSessionTTL).First(&info).Error; err == nil {
		if n, convErr := strconv.Atoi(info.Value); convErr == nil && n > 0 {
			ttlSec = n
		}
	}

	expires := time.Now().Add(time.Duration(ttlSec) * time.Second)
	result := p.db.WithContext(p.ctx).
		Model(&entity.PreviewSession{}).
		Where("expires_at IS NULL").
		Update("expires_at", expires)
	if result.Error != nil {
		p.log.Warn(p.ctx, fmt.Sprintf("回填 Preview expires_at 失败: %s", result.Error.Error()))
		return
	}
	if result.RowsAffected > 0 {
		p.log.Info(p.ctx, fmt.Sprintf("已为 %d 个历史 Preview 会话回填过期时间", result.RowsAffected))
	}
}
