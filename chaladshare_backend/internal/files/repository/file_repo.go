package repository

import (
	"database/sql"
	"fmt"
	"time"

	"chaladshare_backend/internal/files/models"
)

type FileRepository interface {
	// documents
	CreateDocument(doc *models.Document) (*models.Document, error)
	GetListDocByUserID(userID int) ([]models.Document, error)
	DeleteDocument(id int) error

	GetDocumentOwnerID(documentID int) (int, error)
	GetDocumentByID(documentID int) (*models.Document, error)

	// summaries
	GetSummaryByDocID(docID int) (*models.Summary, error)
	CreateSummary(summary *models.Summary) (*models.Summary, error)
	DeleteSummariesByDocID(docID int) error
}

type fileRepository struct {
	db *sql.DB
}

func NewFileRepository(db *sql.DB) FileRepository {
	return &fileRepository{db: db}
}

// CreateDocument
func (r *fileRepository) CreateDocument(req *models.Document) (*models.Document, error) {
	err := r.db.QueryRow(`
		INSERT INTO documents (document_user_id, document_name, document_url, storage_provider, uploaded_at)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING document_id, uploaded_at
	`,
		req.DocumentUserID, req.DocumentName, req.DocumentURL, req.StorageProvider, time.Now(),
	).Scan(&req.DocumentID, &req.UploadedAt)

	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถบันทึกไฟล์ได้: %v", err)
	}
	return req, nil
}

// etListDocByUserID latest
func (r *fileRepository) GetListDocByUserID(userID int) ([]models.Document, error) {
	rows, err := r.db.Query(`
		SELECT document_id, document_user_id, document_name, document_url, storage_provider, uploaded_at
		FROM documents
		WHERE document_user_id = $1
		ORDER BY uploaded_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []models.Document
	for rows.Next() {
		var d models.Document
		if err := rows.Scan(&d.DocumentID, &d.DocumentUserID, &d.DocumentName, &d.DocumentURL, &d.StorageProvider, &d.UploadedAt); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, nil
}

// CreateSummary
func (r *fileRepository) CreateSummary(summary *models.Summary) (*models.Summary, error) {
	err := r.db.QueryRow(`
		INSERT INTO summaries (summary_text, summary_html, summary_pdf_url, summary_created_at, document_id)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING summary_id, summary_created_at
	`, summary.SummaryText, summary.SummaryHTML, summary.SummaryPDFURL, time.Now(), summary.DocumentID).
		Scan(&summary.SummaryID, &summary.SummaryCreatedAt)
	if err != nil {
		return nil, err
	}
	return summary, nil
}

// GetSummaryByDocID
func (r *fileRepository) GetSummaryByDocID(docID int) (*models.Summary, error) {
	var s models.Summary
	err := r.db.QueryRow(`
		SELECT summary_id, summary_text, summary_html, summary_pdf_url, summary_created_at, document_id
		FROM summaries
		WHERE document_id = $1
	`, docID).Scan(&s.SummaryID, &s.SummaryText, &s.SummaryHTML, &s.SummaryPDFURL, &s.SummaryCreatedAt, &s.DocumentID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("ไม่พบสรุปของเอกสารนี้")
	}
	return &s, err
}

// DeleteDocument
func (r *fileRepository) DeleteDocument(id int) error {
	res, err := r.db.Exec("DELETE FROM documents WHERE document_id = $1", id)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *fileRepository) GetDocumentOwnerID(documentID int) (int, error) {
	var ownerID int
	err := r.db.QueryRow(`
        SELECT document_user_id
        FROM documents
        WHERE document_id = $1
    `, documentID).Scan(&ownerID)
	if err != nil {
		return 0, err
	}
	return ownerID, nil
}

func (r *fileRepository) GetDocumentByID(id int) (*models.Document, error) {
	var d models.Document
	err := r.db.QueryRow(
		`SELECT document_id, document_user_id, document_name, document_url, storage_provider, uploaded_at
		FROM documents
		WHERE document_id = $1`, id).Scan(&d.DocumentID, &d.DocumentUserID, &d.DocumentName, &d.DocumentURL, &d.StorageProvider, &d.UploadedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *fileRepository) DeleteSummariesByDocID(docID int) error {
	_, err := r.db.Exec(`DELETE FROM summaries WHERE document_id = $1`, docID)
	return err
}
