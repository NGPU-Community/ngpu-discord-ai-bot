package header

type RequestSadTalker struct {
	ImageUrl       string `json:"image_url"`
	Text           string `json:"text"`
	Pronouncer     string `json:"pronouncer"`
	BackGroundName string `json:"backGroundName"`
	LogoUrl        string `json:"logo_url"`
}

type ResponseSadTalker struct {
	ResultCode int    `json:"result_code"`
	Message    string `json:"message"`
	TaskId     string `json:"taskId"`
}
