package header

type RequestTxt2Img struct {
	Prompt string `json:"prompt"`
	Width  int    `json:"width" binding:"omitempty"`
	Height int    `json:"height" binding:"omitempty"`
}
