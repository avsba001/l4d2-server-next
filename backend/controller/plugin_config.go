package controller

import (
	"l4d2-manager-next/logic"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetPluginConfig(c *gin.Context) {
	pluginName := c.PostForm("name")
	if pluginName == "" {
		c.String(http.StatusBadRequest, "插件名称不能为空")
		return
	}

	configs, err := logic.GetPluginConfigs(pluginName)
	if err != nil {
		c.String(http.StatusInternalServerError, "获取插件配置失败: %v", err)
		return
	}

	c.JSON(http.StatusOK, configs)
}

type UpdateConfigRequest struct {
	ConfigName string            `json:"config_name"`
	Updates    map[string]string `json:"updates"`
}

func UpdatePluginConfig(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "admin" {
		c.String(http.StatusForbidden, "需要管理员权限")
		return
	}

	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, "无效的请求格式")
		return
	}

	if err := logic.SavePluginConfig(req.ConfigName, req.Updates); err != nil {
		c.String(http.StatusInternalServerError, "保存配置失败: %v", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}
