package types

//TravisCI `/requests` API structure response
type TravisRespStruct struct {
	Requests []struct {
		Builds []struct {
			Href              string `json:"@href"`
			PullRequestNumber int    `json:"pull_request_number"`
		} `json:"builds"`
	} `json:"requests"`
}

//TravisCI `/jobs` API structure response
type TravisJobRespStruct struct {
	Jobs []struct {
		Type   string `json:"@type"`
		Href   string `json:"@href"`
		Number string `json:"number"`
	} `json:"jobs"`
}

