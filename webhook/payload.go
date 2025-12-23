package webhook

type Payload struct {
	Content string `json:"content"`
	UserName string `json:"username,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}