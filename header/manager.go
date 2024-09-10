package header

/*********************************/

type UserKey struct {
	DiscordId  string `json:"discordId"`
	UserKey    string `json:"userKey"`
	CreateTime string `json:"createTime"`
	SpareTime  int64  `json:"spareTime"`
}

type ResponseRegister struct {
	ResultCode int64  `json:"result_code"` //返回消息
	Msg        string `json:"msg"`         //返回消息
	Key        string `json:"key"`         //用于验证客户信息的字符串
	KeyLen     int64  `json:"len"`
}

type ResponseUserInfo struct {
	ResultCode    int64  `json:"result_code"` //返回消息
	Msg           string `json:"msg"`         //返回消息
	DiscordId     string `json:"discordId"`
	UserKey       string `json:"userKey"`
	EMail         string `json:"eMail"`
	RemainingTime int64  `json:"remainingTime"`
	Subscribed    int64  `json:"subscribed"`
	CreateTime    string `json:"createTime"`
	Lasttime      string `json:"lasttime"`
	State         int64  `json:"state"`
}
