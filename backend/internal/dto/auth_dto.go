package dto

type LoginRequest struct {
	EmployeeID string `json:"employee_id" validate:"required"`
	Password   string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	User         UserInfo `json:"user"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}

type UserInfo struct {
	ID            string `json:"id"`
	EmployeeID    string `json:"employee_id"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	PriorityLevel int    `json:"priority_level"`
}

type UpdateFCMTokenRequest struct {
	Token string `json:"token" validate:"required"`
}
