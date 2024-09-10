package header

type RequestBlackWhite2Color struct {
	Fun      string  `json:"fun"`
	Source   string  `json:"source"`
	Strength float32 `json:"strength"`
}

type ResponseBlackWhite2Color struct {
	ResultCode int    `json:"result_code"`
	Message    string `json:"message"`
	TaskId     string `json:"taskId"`
}

type RequestLipSyncer struct {
	Fun       string `json:"fun"`
	Audio     string `json:"audio"`
	Video     string `json:"video"`
	IsEnhance bool   `json:"isEnhance"`
}

type ResponseLipSyncer struct {
	ResultCode int    `json:"result_code"`
	Message    string `json:"message"`
	TaskId     string `json:"taskId"`
}

type RequestFaceSwap struct {
	Fun                 string  `json:"fun"`
	FaceReference       string  `json:"faceReference"`
	FaceTarget          string  `json:"faceTarget"`
	Strength            float32 `json:"strength"`
	IsFaceEnhance       bool    `json:"isFaceEnhance"`
	FaceEnhanceStrength float32 `json:"faceEnhanceStrength"`
}

type ResponseFaceSwap struct {
	ResultCode int    `json:"result_code"`
	Message    string `json:"message"`
	TaskId     string `json:"taskId"`
}

type RequestFrameEnhance struct {
	Fun                 string  `json:"fun"`
	Source              string  `json:"source"`
	Strength            float32 `json:"strength"`
	IsFaceEnhance       bool    `json:"isFaceEnhance"`
	FaceEnhanceStrength float32 `json:"faceEnhanceStrength"`
}

type ResponseFrameEnhance struct {
	ResultCode int    `json:"result_code"`
	Message    string `json:"message"`
	TaskId     string `json:"taskId"`
}

type FaceFusionDBData struct {
	TaskID       string `json:"taskId"`
	Method       string `json:"method"`
	RequestTime  string `json:"requestTime"`
	ResponseData string `json:"responseData"`
}
