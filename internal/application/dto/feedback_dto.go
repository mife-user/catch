package dto

type FeedbackRequest struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type FeedbackResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type SMTPTemplate struct {
	Name string `json:"name"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

type SMTPTestRequest struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	To       string `json:"to"`
}

type SMTPTestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
