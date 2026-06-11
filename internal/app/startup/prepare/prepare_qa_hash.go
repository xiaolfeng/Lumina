package prepare

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/xiaolfeng/Lumina/internal/entity"
)

// prepareQaHash 幂等回填已有 Session 的 Hash 字段
func (p *Prepare) prepareQaHash() {
	var sessions []entity.QaSession
	if err := p.db.WithContext(p.ctx).Where("hash = ?", "").Find(&sessions).Error; err != nil {
		p.log.Warn(p.ctx, fmt.Sprintf("查询空Hash会话失败: %s", err.Error()))
		return
	}
	if len(sessions) == 0 {
		return
	}

	p.log.Info(p.ctx, fmt.Sprintf("发现 %d 个会话需要回填Hash", len(sessions)))
	for _, s := range sessions {
		sum := sha256.Sum256([]byte(s.ID.String()))
		h := hex.EncodeToString(sum[:])[:16]
		if err := p.db.WithContext(p.ctx).Model(&s).Update("hash", h).Error; err != nil {
			p.log.Warn(p.ctx, fmt.Sprintf("回填Hash失败 [%s]: %s", s.ID.String(), err.Error()))
		}
	}
}
