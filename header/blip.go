package header

type BaseImg2Txt struct {
	Task  string `json:"task"`
	Image string `json:"image"`
}

type RequestImg2Txt struct {
	Input BaseImg2Txt `json:"input"`
}

type ResponseImg2Txt struct {
	ResultCode  int    `json:"result_code"`
	Message     string `json:"message"`
	TaskDursion int64  `json:"taskDursion"`
	Data        string `json:"data"`
}
