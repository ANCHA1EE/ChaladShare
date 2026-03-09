package models

type SeedItem struct {
	DocumentID     int       `json:"document_id"`
	StyleLabel     string    `json:"style_label"`
	ClusterID      int       `json:"cluster_id"`
	StyleVectorV16 []float64 `json:"style_vector_v16"`
}

type CandidateItem struct {
	PostID         int       `json:"post_id"`
	DocumentID     int       `json:"document_id"`
	StyleLabel     string    `json:"style_label"`
	ClusterID      int       `json:"cluster_id"`
	StyleVectorV16 []float64 `json:"style_vector_v16"`
}

type ColabRecommendFromLikedReq struct {
	Seeds            []SeedItem      `json:"seeds"`
	Candidates       []CandidateItem `json:"candidates"`
	TopK             int             `json:"top_k"`
	BoostSameCluster float64         `json:"boost_same_cluster"`
	MaxPerCluster    int             `json:"max_per_cluster"`
}

type ColabRecommendItem struct {
	PostID                int     `json:"post_id"`
	DocumentID            int     `json:"document_id"`
	StyleLabel            string  `json:"style_label"`
	ClusterID             int     `json:"cluster_id"`
	Score                 float64 `json:"score"`
	MatchedSeedDocumentID *int    `json:"matched_seed_document_id,omitempty"`
}

type ColabRecommendFromLikedResp struct {
	Recommendations []ColabRecommendItem `json:"recommendations"`
	Meta            map[string]any       `json:"meta,omitempty"`
}
