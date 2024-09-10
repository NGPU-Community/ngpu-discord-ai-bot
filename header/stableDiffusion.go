package header

type ResponseStableDiffusion struct {
	ResultCode  int    `json:"result_code"`
	Message     string `json:"message"`
	TaskDursion int64  `json:"taskDursion"`
	Data        []byte `json:"data"`
}
