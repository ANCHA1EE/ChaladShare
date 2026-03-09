package models

type VectorItem struct {
	DocumentID     int       `json:"document_id"`
	StyleLabel     string    `json:"style_label"`
	StyleVectorV16 []float64 `json:"style_vector_v16"`
}

type VectorsResp struct {
	Items []VectorItem `json:"items"`
}

type ClusterUpdate struct {
	DocumentID int `json:"document_id"`
	ClusterID  int `json:"cluster_id"`
}

type BatchUpdateClustersReq struct {
	Updates []ClusterUpdate `json:"updates"`
}

type BatchUpdateClustersResp struct {
	Updated int `json:"updated"`
}
