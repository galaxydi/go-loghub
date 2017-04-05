package sls

// GetLogsResponse defines response from GetLogs call
type GetLogsResponse struct {
	Progress string              `json:"progress"`
	Count    int64               `json:"count"`
	Logs     []map[string]string `json:"logs"`
}

// IndexKey ...
type IndexKey struct {
	Tokens        []string // tokens that split the log line.
	CaseSensitive bool
	Type          string // text, long, double
}

type Index struct {
}
