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
