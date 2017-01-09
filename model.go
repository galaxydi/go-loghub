package sls

// GetLogsResponse defines response from GetLogs call
type GetLogsResponse struct {
	Progress string              `json:"progress"`
	Count    int64               `json:"count"`
	Logs     []map[string]string `json:"logs"`
}
