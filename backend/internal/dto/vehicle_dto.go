package dto

type ToggleMaintenanceRequest struct {
	IsMaintenance bool `json:"is_maintenance"`
}

type CreateVehicleRequest struct {
	Name         string `json:"name"`
	LicensePlate string `json:"license_plate"`
	DriverID     string `json:"driver_id"`
}

type UpdateVehicleRequest struct {
	Name         string `json:"name"`
	LicensePlate string `json:"license_plate"`
	DriverID     string `json:"driver_id"`
}
