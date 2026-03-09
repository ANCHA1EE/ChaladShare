package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"chaladshare_backend/internal/files/models"
	"chaladshare_backend/internal/files/service"
	"chaladshare_backend/internal/middleware"
)

type FileHandler struct {
	fileservice service.FileService
}

func NewFileHandler(fileservice service.FileService) *FileHandler {
	return &FileHandler{fileservice: fileservice}
}

// File supabase
func (h *FileHandler) UploadFile(c *gin.Context) {
	uid := c.GetInt(middleware.CtxUserID)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาแนบไฟล์ PDF"})
		return
	}

	id := uuid.New().String()
	filename := id + ".pdf"

	abs := filepath.Join(os.TempDir(), filename)

	if err := c.SaveUploadedFile(fh, abs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกไฟล์ได้"})
		return
	}

	req := &models.UploadRequest{
		UserID:          uid,
		DocumentName:    fh.Filename,
		DocumentURL:     "",
		StorageProvider: "supabase",
		LocalPath:       abs,
	}
	resp, err := h.fileservice.UploadFile(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"document_id": resp.DocumentID,
		"pdf_url":     resp.FileURL,
	})
}

// cover supabase
func (h *FileHandler) UploadCover(c *gin.Context) {
	uid := c.GetInt(middleware.CtxUserID)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาแนบรูปหน้าปก"})
		return
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รองรับเฉพาะ .jpg .jpeg .png"})
		return
	}

	id := uuid.New().String()
	filename := fmt.Sprintf("cover_%s_%d%s", id, time.Now().UnixNano(), ext)
	abs := filepath.Join(os.TempDir(), filename)

	if err := c.SaveUploadedFile(fh, abs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกไฟล์หน้าปกได้"})
		return
	}

	st, err := service.NewSupabaseStorageFromEnv()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	objectPath := fmt.Sprintf("covers/%d/%s", uid, filename)
	publicURL, err := st.UploadLocalFile(c.Request.Context(), objectPath, abs)
	_ = os.Remove(abs)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"cover_url":     publicURL,
		"cover_storage": "supabase",
	})
}

// Avatar supabase
func (h *FileHandler) UploadAvatar(c *gin.Context) {
	uid := c.GetInt(middleware.CtxUserID)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาแนบรูปโปรไฟล์"})
		return
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รองรับเฉพาะ .jpg .jpeg .png"})
		return
	}

	id := uuid.New().String()
	filename := fmt.Sprintf("avatar_%d_%s_%d%s", uid, id, time.Now().UnixNano(), ext)
	abs := filepath.Join(os.TempDir(), filename)

	if err := c.SaveUploadedFile(fh, abs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกรูปโปรไฟล์ได้"})
		return
	}
	defer func() { _ = os.Remove(abs) }()

	st, err := service.NewSupabaseStorageFromEnv()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	objectPath := fmt.Sprintf("avatars/%d/%s", uid, filename)
	publicURL, err := st.UploadLocalFile(c.Request.Context(), objectPath, abs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"avatar_url":     publicURL,
		"avatar_storage": "supabase",
	})
}

// file local
// func (h *FileHandler) UploadFile(c *gin.Context) {
// 	uid := c.GetInt(middleware.CtxUserID)
// 	if uid == 0 {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
// 		return
// 	}

// 	fh, err := c.FormFile("file")
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาแนบไฟล์ PDF"})
// 		return
// 	}

// 	if err := os.MkdirAll("./uploads", 0755); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถสร้างโฟลเดอร์ uploads ได้"})
// 		return
// 	}

// 	id := uuid.New().String()
// 	filename := id + ".pdf"
// 	abs := filepath.Join("./uploads", filename)
// 	publicURL := "/uploads/" + filename

// 	if err := c.SaveUploadedFile(fh, abs); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกไฟล์ได้"})
// 		return
// 	}

// 	req := &models.UploadRequest{
// 		UserID:          uid,
// 		DocumentName:    fh.Filename,
// 		DocumentURL:     publicURL,
// 		StorageProvider: "local",
// 		LocalPath:       abs,
// 	}
// 	resp, err := h.fileservice.UploadFile(req)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, gin.H{
// 		"document_id": resp.DocumentID,
// 		"pdf_url":     publicURL,
// 	})
// }

// // cover image local
// func (h *FileHandler) UploadCover(c *gin.Context) {
// 	uid := c.GetInt(middleware.CtxUserID)
// 	if uid == 0 {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
// 		return
// 	}

// 	fh, err := c.FormFile("file")
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาแนบรูปหน้าปก"})
// 		return
// 	}

// 	ext := strings.ToLower(filepath.Ext(fh.Filename))
// 	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "รองรับเฉพาะ .jpg .jpeg .png"})
// 		return
// 	}

// 	baseDir := "./uploads/covers"
// 	if err := os.MkdirAll(baseDir, 0755); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถสร้างโฟลเดอร์ได้"})
// 		return
// 	}

// 	id := uuid.New().String()
// 	filename := fmt.Sprintf("cover_%s_%d%s", id, time.Now().UnixNano(), ext)
// 	abs := filepath.Join(baseDir, filename)

// 	publicURL := "/uploads/covers/" + filename

// 	if err := c.SaveUploadedFile(fh, abs); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกไฟล์หน้าปกได้"})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, gin.H{
// 		"cover_url": publicURL,
// 	})
// }

// // avatar image local
// func (h *FileHandler) UploadAvatar(c *gin.Context) {
// 	uid := c.GetInt(middleware.CtxUserID)
// 	if uid == 0 {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
// 		return
// 	}

// 	fh, err := c.FormFile("file")
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาแนบรูปโปรไฟล์"})
// 		return
// 	}

// 	ext := strings.ToLower(filepath.Ext(fh.Filename)) // .jpg / .png ...
// 	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "รองรับเฉพาะ .jpg .jpeg .png"})
// 		return
// 	}

// 	baseDir := "./uploads/avatars"
// 	if err := os.MkdirAll(baseDir, 0755); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถสร้างโฟลเดอร์ได้"})
// 		return
// 	}

// 	id := uuid.New().String()
// 	filename := fmt.Sprintf("avatar_%d_%s_%d%s", uid, id, time.Now().UnixNano(), ext)
// 	abs := filepath.Join(baseDir, filename)

// 	publicURL := "/uploads/avatars/" + filename

// 	if err := c.SaveUploadedFile(fh, abs); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกรูปโปรไฟล์ได้"})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, gin.H{
// 		"avatar_url":     publicURL,
// 		"avatar_storage": "local",
// 	})
// }

// GET
func (h *FileHandler) GetFilesByUserID(c *gin.Context) {
	authUID := c.GetInt(middleware.CtxUserID)
	targetUID, err := strconv.Atoi(c.Param("id"))
	if err != nil || targetUID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}
	if authUID != targetUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	files, err := h.fileservice.GetFilesByUserID(targetUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบไฟล์ของผู้ใช้นี้"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, files)
}

// GET /api/v1/documents/:document_id/summary
func (h *FileHandler) GetSummaryByDocumentID(c *gin.Context) {
	authUID := c.GetInt(middleware.CtxUserID)
	docID, err := strconv.Atoi(c.Param("document_id"))
	if err != nil || docID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document_id"})
		return
	}

	ok, err := h.fileservice.IsOwner(docID, authUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	summary, err := h.fileservice.GetSummaryByDocumentID(docID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบสรุปของไฟล์นี้"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

// DELETE
func (h *FileHandler) DeleteFile(c *gin.Context) {
	authUID := c.GetInt(middleware.CtxUserID)
	if authUID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	docID, err := strconv.Atoi(c.Param("document_id"))
	if err != nil || docID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document_id"})
		return
	}

	ok, err := h.fileservice.IsOwner(docID, authUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	if err := h.fileservice.DeleteFile(docID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบไฟล์นี้"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ลบไฟล์สำเร็จ"})
}
