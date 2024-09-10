package header

type RequestRemoveBG struct {
	Kind     string `json:"kind"`
	Obj      string `json:"obj"`
	ImageUrl string `json:"url"`
	BGColor  string `json:"bgColor"`
}

type ResponseRemoveBG struct {
	ResultCode  int    `json:"result_code"`
	Message     string `json:"message"`
	TaskDursion int64  `json:"taskDursion"`
	Data        string `json:"data"`
}

type RequestReplaceBG struct {
	Kind    string `json:"kind"`
	Obj     string `json:"obj"`
	SrcUrl  string `json:"url"`
	BGUrl   string `json:"bgPhoto"`
	BGColor string `json:"bgColor"`
}

type ResponseReplaceBG struct {
	ResultCode  int    `json:"result_code"`
	Message     string `json:"message"`
	TaskDursion int64  `json:"taskDursion"`
	Data        string `json:"data"`
}
