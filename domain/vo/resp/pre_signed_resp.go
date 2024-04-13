package resp

type PreSignedResp struct {
	Url    string            `json:"url"`
	Policy map[string]string `json:"policy"`
}
