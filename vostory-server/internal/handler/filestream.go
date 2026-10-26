package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// streamAttachment 把存储层打开的文件流写回响应
func streamAttachment(ctx *gin.Context, r io.ReadCloser, contentType, disposition string) {
	defer r.Close()
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	ctx.Header("Content-Type", contentType)
	if disposition != "" {
		ctx.Header("Content-Disposition", disposition)
	}
	ctx.Status(http.StatusOK)
	_, _ = io.Copy(ctx.Writer, r)
}
