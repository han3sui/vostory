package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsFileImportHandler struct {
	*Handler
	svc service.VsFileImportService
}

func NewVsFileImportHandler(handler *Handler, svc service.VsFileImportService) *VsFileImportHandler {
	return &VsFileImportHandler{Handler: handler, svc: svc}
}

// Upload godoc
// @Summary      上传源文件
// @Description  为项目上传 txt/docx/epub 源文件
// @Tags         文件导入
// @Accept       multipart/form-data
// @Produce      json
// @Param        project_id  formData  int   true  "项目ID"
// @Param        file        formData  file  true  "源文件"
// @Success      200         {object}  v1.Response
// @Failure      400         {object}  v1.Response
// @Failure      500         {object}  v1.Response
// @Router       /api/v1/project/import/upload [post]
// @Id        project:import:upload
func (h *VsFileImportHandler) Upload(ctx *gin.Context) {
	projectID := cast.ToUint64(ctx.PostForm("project_id"))
	if projectID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "project_id is required"), nil)
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "file is required"), nil)
		return
	}

	sourceType, filePath, err := h.svc.UploadFile(ctx, projectID, file)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, &v1.VsFileImportResponse{
		ProjectID:     projectID,
		FileName:      file.Filename,
		FileSize:      file.Size,
		SourceType:    sourceType,
		SourceFileURL: filePath,
	})
}

