package controller

import (
	"fmt"
	"l4d2-manager-next/logic"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetPlugins(c *gin.Context) {
	plugins, err := logic.GetPlugins()
	if err != nil {
		c.String(http.StatusInternalServerError, "获取插件列表失败: %v", err)
		return
	}
	c.JSON(http.StatusOK, plugins)
}

func UploadPlugin(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "admin" {
		c.String(http.StatusForbidden, "需要管理员权限")
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		if err = c.Request.ParseMultipartForm(32 << 20); err != nil {
			c.String(http.StatusBadRequest, "解析表单失败: %v", err)
			return
		}
		form = c.Request.MultipartForm
	}

	files := form.File["file"]
	if len(files) == 0 {
		c.String(http.StatusBadRequest, "未上传文件")
		return
	}

	var errs []string
	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s 打开失败: %v", header.Filename, err))
			continue
		}

		if err := logic.UploadPlugin(file, header.Size, header.Filename); err != nil {
			errs = append(errs, fmt.Sprintf("%s 上传失败: %v", header.Filename, err))
		}
		file.Close()
	}

	if len(errs) > 0 {
		c.String(http.StatusInternalServerError, strings.Join(errs, "; "))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "插件上传成功"})
}

func EnablePlugin(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "admin" {
		c.String(http.StatusForbidden, "需要管理员权限")
		return
	}

	name := c.PostForm("name")
	if name == "" {
		c.String(http.StatusBadRequest, "插件名称不能为空")
		return
	}

	if err := logic.EnablePlugin(name); err != nil {
		c.String(http.StatusInternalServerError, "启用插件失败: %v", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "插件启用成功"})
}

func DisablePlugin(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "admin" {
		c.String(http.StatusForbidden, "需要管理员权限")
		return
	}

	name := c.PostForm("name")
	if name == "" {
		c.String(http.StatusBadRequest, "插件名称不能为空")
		return
	}

	if err := logic.DisablePlugin(name); err != nil {
		c.String(http.StatusInternalServerError, "禁用插件失败: %v", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "插件禁用成功"})
}

func DeletePlugin(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "admin" {
		c.String(http.StatusForbidden, "需要管理员权限")
		return
	}

	name := c.PostForm("name")
	if name == "" {
		c.String(http.StatusBadRequest, "插件名称不能为空")
		return
	}

	if err := logic.DeletePlugin(name); err != nil {
		c.String(http.StatusInternalServerError, "删除插件失败: %v", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "插件删除成功"})
}

type BatchPluginRequest struct {
	Names []string `json:"names"`
}

func EnablePlugins(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "admin" {
		c.String(http.StatusForbidden, "需要管理员权限")
		return
	}

	var req BatchPluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, "无效的请求格式")
		return
	}

	if len(req.Names) == 0 {
		c.String(http.StatusBadRequest, "插件列表不能为空")
		return
	}

	if err := logic.EnablePlugins(req.Names); err != nil {
		c.String(http.StatusInternalServerError, "批量启用插件失败: %v", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "批量启用成功"})
}

func DisablePlugins(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "admin" {
		c.String(http.StatusForbidden, "需要管理员权限")
		return
	}

	var req BatchPluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, "无效的请求格式")
		return
	}

	if len(req.Names) == 0 {
		c.String(http.StatusBadRequest, "插件列表不能为空")
		return
	}

	if err := logic.DisablePlugins(req.Names); err != nil {
		c.String(http.StatusInternalServerError, "批量禁用插件失败: %v", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "批量禁用成功"})
}

func GetPresets(c *gin.Context) {
	presets, err := logic.GetPresets()
	if err != nil {
		c.String(http.StatusInternalServerError, "获取预设列表失败: %v", err)
		return
	}
	c.JSON(http.StatusOK, presets)
}

func ApplyPreset(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "admin" {
		c.String(http.StatusForbidden, "需要管理员权限")
		return
	}

	name := c.PostForm("name")
	if name == "" {
		c.String(http.StatusBadRequest, "预设名称不能为空")
		return
	}

	if err := logic.ApplyPreset(name); err != nil {
		c.String(http.StatusInternalServerError, "%v", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "预设应用成功"})
}
