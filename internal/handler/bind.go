package handler

import (
	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xVaild "github.com/bamboo-services/bamboo-base-go/common/validator"
	"github.com/gin-gonic/gin"
)

// BindJSON 绑定 JSON 请求体，失败时通过 xVaild.HandleValidationError 返回中文验证错误（40014）。
//
// 返回 true 表示绑定成功，false 表示失败（错误已写入 ctx，调用方直接 return）。
func BindJSON(ctx *gin.Context, obj any) bool {
	if err := ctx.ShouldBindJSON(obj); err != nil {
		xVaild.HandleValidationError(ctx, err)
		return false
	}
	return true
}

// BindQuery 绑定 Query 参数，失败时通过 xVaild.HandleValidationError 返回中文验证错误（40014）。
//
// 返回 true 表示绑定成功，false 表示失败（错误已写入 ctx，调用方直接 return）。
func BindQuery(ctx *gin.Context, obj any) bool {
	if err := ctx.ShouldBindQuery(obj); err != nil {
		xVaild.HandleValidationError(ctx, err)
		return false
	}
	return true
}

// ParseSnowflakeID 从字符串解析雪花 ID，失败时返回 ParameterError（40011/HTTP 400）。
func ParseSnowflakeID(ctx *gin.Context, idStr string) (xSnowflake.SnowflakeID, *xError.Error) {
	id, err := xSnowflake.ParseSnowflakeID(idStr)
	if err != nil {
		return 0, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage("无效的ID: "+err.Error()), false, err)
	}
	return id, nil
}
