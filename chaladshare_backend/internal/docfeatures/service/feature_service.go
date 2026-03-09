package service

import (
	"fmt"
	"log"

	"chaladshare_backend/internal/connect"
	"chaladshare_backend/internal/docfeatures/models"
	"chaladshare_backend/internal/docfeatures/repository"
)

type FeatureService interface {
	CreateQueued(documentID int) error
	MarkProcessing(documentID int) error
	SaveResult(input models.SaveResult) error
	MarkFailed(documentID int, msg string) error
	GetByDocumentID(documentID int) (*models.DocumentFeature, error)
	ProcessDocument(documentID int, pdfPath string)

	//
	ListVectors(label string, onlyUnclustered bool) ([]models.VectorItem, error)
	BatchUpdateClusters(updates []models.ClusterUpdate) (int, error)
	RunClustering(label string, onlyUnclustered bool, k int) (int, error)
	BootstrapAutoClustering()
	DeleteByDocumentID(documentID int) error
}

type featureService struct {
	featureRepo repository.DocFeaturesRepo
	aiClient    *connect.Client
}

func NewFeatureService(featureRepo repository.DocFeaturesRepo, aiClient *connect.Client) FeatureService {
	return &featureService{
		featureRepo: featureRepo,
		aiClient:    aiClient,
	}
}

func (s *featureService) CreateQueued(documentID int) error {
	if documentID <= 0 {
		return fmt.Errorf("invalid documentID")
	}
	return s.featureRepo.CreateQueued(documentID)
}

func (s *featureService) MarkProcessing(documentID int) error {
	if documentID <= 0 {
		return fmt.Errorf("invalid documentID")
	}
	return s.featureRepo.MarkProcessing(documentID)
}

func (s *featureService) SaveResult(input models.SaveResult) error {
	if input.DocumentID <= 0 {
		return fmt.Errorf("invalid documentID")
	}
	return s.featureRepo.SaveResult(input)
}

func (s *featureService) MarkFailed(documentID int, msg string) error {
	if documentID <= 0 {
		return fmt.Errorf("invalid documentID")
	}
	if msg == "" {
		msg = "unknown error"
	}
	return s.featureRepo.MarkFailed(documentID, msg)
}

func (s *featureService) GetByDocumentID(documentID int) (*models.DocumentFeature, error) {
	if documentID <= 0 {
		return nil, fmt.Errorf("invalid documentID")
	}
	return s.featureRepo.GetByDocumentID(documentID)
}

func (s *featureService) ProcessDocument(documentID int, pdfPath string) {
	if s.aiClient == nil {
		_ = s.MarkFailed(documentID, "ai client is nil")
		return
	}

	if pdfPath == "" {
		_ = s.MarkFailed(documentID, "pdfPath is empty")
		return
	}

	if err := s.MarkProcessing(documentID); err != nil {
		_ = s.MarkFailed(documentID, err.Error())
		return
	}

	resp, err := s.aiClient.ExtractFeatures(documentID, pdfPath)
	if err != nil {
		_ = s.MarkFailed(documentID, err.Error())
		return
	}

	if resp.StyleLabel == nil || *resp.StyleLabel == "" {
		_ = s.MarkFailed(documentID, "missing style label ")
		return
	}

	if len(resp.StyleVectorV16) == 0 {
		_ = s.MarkFailed(documentID, "empty style_vector_v16 from ai")
		return
	}

	label := *resp.StyleLabel
	ct := resp.ContentText
	if err := s.SaveResult(models.SaveResult{
		DocumentID:       documentID,
		StyleLabel:       label,
		StyleVectorV16:   resp.StyleVectorV16,
		ContentText:      &ct,
		ContentEmbedding: resp.Embedding,
		ClusterID:        resp.ClusterID,
	}); err != nil {
		_ = s.MarkFailed(documentID, err.Error())
		return
	}
	if label == "typed" || label == "handwritten" {
		go s.autoClusterIfReady(label)
		log.Printf("[AUTO-CLUSTER] trigger from ProcessDocument label=%s", label)
	}
}

func (s *featureService) ListVectors(label string, onlyUnclustered bool) ([]models.VectorItem, error) {
	if label == "" {
		return nil, fmt.Errorf("label is empty")
	}
	return s.featureRepo.ListVectors(label, onlyUnclustered)
}

func (s *featureService) BatchUpdateClusters(updates []models.ClusterUpdate) (int, error) {
	return s.featureRepo.BatchUpdateClusters(updates)
}

func (s *featureService) RunClustering(label string, onlyUnclustered bool, k int) (int, error) {
	if s.aiClient == nil {
		return 0, fmt.Errorf("ai client is nil")
	}
	if label != "typed" && label != "handwritten" {
		return 0, fmt.Errorf("label must be typed or handwritten")
	}
	if k <= 0 {
		k = 8
	}

	items, err := s.featureRepo.ListVectors(label, onlyUnclustered)
	if err != nil {
		return 0, err
	}
	if len(items) < 2 {
		// น้อยกว่า 2 ทำ cluster ไม่ meaningful
		return 0, nil
	}

	colabResp, err := s.aiClient.ClusterBatch(models.ColabClusterReq{
		Items:        items,
		K:            k,
		UseScaler:    false,
		SplitByLabel: false,
		RandomState:  42,
	})
	if err != nil {
		return 0, err
	}

	return s.featureRepo.BatchUpdateClusters(colabResp.Updates)
}

func (s *featureService) BootstrapAutoClustering() {
	// รันตอน start เพื่อจัดการไฟล์ค้างที่ยัง cluster_id = NULL
	go s.autoClusterIfReady("typed")
	go s.autoClusterIfReady("handwritten")
}

func (s *featureService) autoClusterIfReady(label string) {
	nNew, err := s.featureRepo.CountUnclustered(label)
	if err != nil {
		log.Printf("[AUTO-CLUSTER] count NEW error label=%s: %v", label, err)
		return
	}
	log.Printf("[AUTO-CLUSTER] label=%s unclustered(new)=%d", label, nNew)

	if nNew < 10 {
		return
	}

	nAll, err := s.featureRepo.CountClusterable(label)
	if err != nil {
		log.Printf("[AUTO-CLUSTER] count ALL error label=%s: %v", label, err)
		return
	}
	log.Printf("[AUTO-CLUSTER] label=%s total(all)=%d", label, nAll)

	// เลือก k จาก “ทั้งหมด”
	k := 4
	if nAll >= 20 {
		k = 5
	}
	if nAll >= 30 {
		k = 6
	}
	log.Printf("[AUTO-CLUSTER] RUN label=%s k=%d (recluster ALL)", label, k)

	updated, err := s.RunClustering(label, false, k) // false = recluster ทั้งชุด
	if err != nil {
		log.Printf("[AUTO-CLUSTER] run error label=%s: %v", label, err)
		return
	}
	log.Printf("[AUTO-CLUSTER] DONE label=%s updated=%d", label, updated)
}

func (s *featureService) DeleteByDocumentID(documentID int) error {
	if documentID <= 0 {
		return fmt.Errorf("invalid documentID")
	}
	return s.featureRepo.DeleteByDocumentID(documentID)
}
