package form

// CostExplorerParams ...
type CostExplorerParams struct {
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
	Granularity string    `json:"granularity"`
	Metric      []*string `json:"metric"`
	Services    []*string `json:"services"`
	GroupName   string    `json:"group_name"`
}

// CostExplorerForcastParams ...
type CostExplorerForcastParams struct {
	EndDate     string `json:"end_date"`
	Granularity string `json:"granularity"`
	Metric      string `json:"metric"`
	StartDate   string `json:"start_date"`
}
