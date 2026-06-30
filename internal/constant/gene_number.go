package bConst

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

const (
	GeneProject             xSnowflake.Gene = 32 // 项目基因
	GeneQaSession           xSnowflake.Gene = 33 // QA会话基因
	GeneQaQuestion          xSnowflake.Gene = 34 // QA问题基因
	GeneQaSupplement        xSnowflake.Gene = 35 // QA补充基因
	GeneBiometricCredential xSnowflake.Gene = 36 // 生物特征凭证
	GenePin                 xSnowflake.Gene = 37 // Pin基因
	GeneApikey              xSnowflake.Gene = 38 // API密钥基因
)
