package header

type RequestFengshui struct {
	ImageUrl string `json:"imageUrl"`
	Prompt   string `json:"prompt"`
}

type RequestChat struct {
	Prompt string `json:"prompt"`
}
