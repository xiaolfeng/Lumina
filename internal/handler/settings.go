package handler

import (
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
	apiSettings "github.com/xiaolfeng/Lumina/api/settings"
)

// 确保 apiCommon 包被编译器识别（swag 注释依赖此导入）
var _ = apiCommon.BaseResponse{}

// GetSettings 获取指定分类的系统设置
//
// @Summary     [管理] 获取分类设置
// @Description 根据分类名称获取该分类下所有设置项的键值、类型与描述，未入库项返回默认值
// @Tags        系统设置接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string  true  "Bearer Access Token"
// @Param       category        path      string  true  "设置分类(如 general/qa/repo/security)"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiSettings.CategorySettingsResponse}  "获取成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "无效的设置分类"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     500  {object}  apiCommon.BaseResponse  "服务器内部错误"
// @Router      /api/v1/settings/{category} [GET]
func (h *SettingsHandler) GetSettings(ctx *gin.Context) {
	h.log.Info(ctx, "GetSettings - 获取分类设置")

	category := ctx.Param("category")

	items, xErr := h.service.settingsLogic.GetByCategory(ctx.Request.Context(), category)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	resp := apiSettings.CategorySettingsResponse{
		Category: category,
		Items:    items,
	}

	xResult.SuccessHasData(ctx, "获取设置成功", resp)
}

// UpdateSettings 更新指定分类的系统设置
//
// @Summary     [管理] 更新分类设置
// @Description 批量更新指定分类下的设置项键值，校验分类合法性、Key 归属与值类型后逐项写入
// @Tags        系统设置接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                              true  "Bearer Access Token"
// @Param       category        path      string                              true  "设置分类(如 general/qa/repo/security)"
// @Param       request         body      apiSettings.UpdateCategorySettingsRequest  true  "更新分类设置请求"
// @Success     200  {object}  apiCommon.BaseResponse  "设置更新成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误或设置项不属于该分类"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     500  {object}  apiCommon.BaseResponse  "服务器内部错误"
// @Router      /api/v1/settings/{category} [PUT]
func (h *SettingsHandler) UpdateSettings(ctx *gin.Context) {
	h.log.Info(ctx, "UpdateSettings - 更新分类设置")

	category := ctx.Param("category")

	var req apiSettings.UpdateCategorySettingsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	kvMap := make(map[string]string, len(req.Items))
	for _, item := range req.Items {
		kvMap[item.Key] = item.Value
	}

	if xErr := h.service.settingsLogic.UpdateByCategory(ctx.Request.Context(), category, kvMap); xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "设置更新成功")
}
