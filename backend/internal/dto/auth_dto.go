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

type PassengerRegisterRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	Password    string `json:"password" validate:"required"`
	Name        string `json:"name" validate:"required"`
}

type PassengerLoginRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	Password    string `json:"password" validate:"required"`
}

type PassengerRideRequest struct {
	PickupAddress  string   `json:"pickup_address" validate:"required"`
	PickupLat      float64  `json:"pickup_lat" validate:"required"`
	PickupLng      float64  `json:"pickup_lng" validate:"required"`
	DropoffAddress string   `json:"dropoff_address,omitempty"`
	DropoffLat     *float64 `json:"dropoff_lat,omitempty"`
	DropoffLng     *float64 `json:"dropoff_lng,omitempty"`
	PassengerName  string   `json:"passenger_name,omitempty"`
	PassengerCount int      `json:"passenger_count,omitempty"`
}
