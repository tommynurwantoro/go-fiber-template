package model

type HealthCheck struct {
	Name    string  `json:"name" example:"Postgre"`
	Status  string  `json:"status" example:"success"`
	IsUp    bool    `json:"is_up" example:"true"`
	Message *string `json:"message,omitempty" example:"Postgre is up and running"`
}

type HealthCheckResponse struct {
	Code      int           `json:"code" example:"200"`
	Status    string        `json:"status" example:"success"`
	Message   string        `json:"message" example:"All services healthy"`
	IsHealthy bool          `json:"is_healthy" example:"true"`
	Result    []HealthCheck `json:"result"`
}
