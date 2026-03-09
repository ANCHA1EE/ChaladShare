package models

type ColabClusterReq struct {
	Items        []VectorItem `json:"items"`
	K            int          `json:"k"`
	UseScaler    bool         `json:"use_scaler"`
	SplitByLabel bool         `json:"split_by_label"`
	RandomState  int          `json:"random_state"`
}

type ColabClusterResp struct {
	Updates []ClusterUpdate `json:"updates"`
	Meta    map[string]any  `json:"meta,omitempty"`
}
