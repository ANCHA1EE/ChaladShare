package models

import "time"

// ข้อมูลไฟล์ที่อัปโหลด
type Document struct {
	DocumentID      int       `json:"document_id"`
	DocumentUserID  int       `json:"document_user_id"`
	DocumentName    string    `json:"document_name"`
	DocumentURL     string    `json:"document_url"`
	StorageProvider string    `json:"storage_provider"`
	UploadedAt      time.Time `json:"uploaded_at"`
}

// เก็บข้อมูลจากไฟล์ที่สรุปเนื้อหาด้วย AI
type Summary struct {
	SummaryID        int       `json:"summary_id"`
	SummaryText      string    `json:"summary_text"`
	SummaryHTML      string    `json:"summary_html"`
	SummaryPDFURL    string    `json:"summary_pdf_url"`
	SummaryCreatedAt time.Time `json:"summary_created_at"`
	DocumentID       int       `json:"document_id"`
}

type UploadRequest struct {
	UserID          int    `json:"-"`
	DocumentName    string `json:"document_name"`
	DocumentURL     string `json:"document_url"`
	StorageProvider string `json:"storage_provider"`
	LocalPath       string `json:"-"`
}

// response ตอนอัปโหลดสำเร็จ
type UploadResponse struct {
	Message    string   `json:"message"`
	File       Document `json:"file"`
	FileURL    string   `json:"file_url"`
	DocumentID int      `json:"document_id"`
}
