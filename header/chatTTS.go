package header

type ResponseQueryTask struct {
	ResultCode int64  `json:"result_code"`
	Msg        string `json:"msg"`
}

type FileData struct {
	Path     string      `json:"path"`
	URL      string      `json:"url"`
	Size     interface{} `json:"size"`
	OrigName string      `json:"orig_name"`
	MimeType interface{} `json:"mime_type"`
	IsStream bool        `json:"is_stream"`
	Meta     MetaData    `json:"meta"`
}

type MetaData struct {
	Type string `json:"_type"`
}

type ResponseData struct {
	FileData []FileData `json:"data"`
}

type ChatTTSData struct {
	Content     string  `json:"content"`
	Temperature float32 `json:"temperature"`
	Top_P       float32 `json:"top_p"`
	Top_K       int     `json:"top_k"`
	Audio_seed  int     `json:"audio_seed"`
	Stream_mode bool    `json:"stream_mode"`
	TaskId      string  `json:"taskId"`
}

type ResultRecord struct {
	ResultCode int    `json:"result_code"`
	Message    string `json:"message"`
	AudioSize  int    `json:"audioSize"`
	AudioPath  string `json:"audioPath"`
	AudioUrl   string `json:"audioUrl"`
}

type RequestCustomAudio struct {
	Content     string  `json:"content"`
	Temperature float32 `json:"temperature"`
	Top_P       float32 `json:"top_p"`
	Top_K       int     `json:"top_k"`
	Audio_seed  int     `json:"audio_seed"`
	Stream_mode bool    `json:"stream_mode"`
}

type ResponseCustomAudio struct {
	ResultCode  int    `json:"result_code"`
	Message     string `json:"message"`
	TaskDursion int64  `json:"task_dursion"`
	Data        string `json:"data"`
}

type RequestGenericAudio struct {
	Content string `json:"content"`
}

type ResponseGenericAudio struct {
	ResultCode  int    `json:"result_code"`
	Message     string `json:"message"`
	TaskDursion int64  `json:"task_dursion"`
	Data        string `json:"data"`
}

type ResponseTTS struct {
	AudioUrl string `json:"audio_url"`
}
