package schema

// Heartbeat response object.
type Heartbeat struct {
	Storage    bool `json:"storage"`
	Permission bool `json:"permission"`
	Cache      bool `json:"cache"`
}

// Hello response object.
type Hello struct {
	ProjectName    string `json:"project_name"`
	ProjectDocs    string `json:"project_docs"`
	ProjectVersion string `json:"project_version"`
	HTTPAPIVersion string `json:"http_api_version"`
	URL            string `json:"url"`
	EOS            string `json:"eos,omitempty"`
	Settings       struct {
		BatchMaxRequests int  `json:"batch_max_requests"`
		Readonly         bool `json:"readonly"`
	} `json:"settings"`
	Capabilities interface{} `json:"capabilities"`
}
