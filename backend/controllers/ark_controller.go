package controllers

import (
	"xhw-service/services"
	"xhw-service/utils"

	"github.com/gin-gonic/gin"
)

type ARKController struct {
	service *services.ARKService
}

func NewARKController() *ARKController {
	return &ARKController{
		service: services.NewARKService(),
	}
}

// GenerateProductTitle 根据图片URL或Base64生成商品标题
// @Summary 图片识别生成商品标题
// @Description 使用火山引擎视觉模型识别图片内容，生成中英文商品标题。支持图片URL或Base64编码图片
// @Tags ARK
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body GenerateProductTitleRequest true "请求参数"
// @Success 200 {object} utils.Response{data=GenerateProductTitleResponse}
// @Failure 401 {object} utils.Response
// @Router /api/v1/ark/product-title [post]
func (ac *ARKController) GenerateProductTitle(c *gin.Context) {
	var req services.GenerateProductTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	// 检查参数：image_url 或 image_base64 必须提供一个
	if req.ImageURL == "" && req.ImageBase64 == "" {
		utils.Fail(c, 400, "参数错误: image_url 或 image_base64 必须提供一个")
		return
	}

	result, err := ac.service.GenerateProductTitle(c.Request.Context(), &req)
	if err != nil {
		utils.Fail(c, 500, "生成标题失败: "+err.Error())
		return
	}

	utils.Success(c, result)
}
