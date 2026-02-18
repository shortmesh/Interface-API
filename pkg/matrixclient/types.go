package matrixclient

type StoreCredentialsRequest struct {
	Username    string `json:"username"`
	AccessToken string `json:"access_token"`
	DeviceID    string `json:"device_id"`
}

type StoreCredentialsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type AddDeviceRequest struct {
	Username     string `json:"username"`
	PlatformName string `json:"platform_name"`
}

type AddDeviceResponse struct {
	DeviceID string `json:"device_id"`
	Platform string `json:"platform"`
}

type DeleteDeviceRequest struct {
	Username     string `json:"username"`
	DeviceID     string `json:"device_id"`
	PlatformName string `json:"platform_name"`
}

type DeleteDeviceResponse struct {
	Status string `json:"status"`
}
