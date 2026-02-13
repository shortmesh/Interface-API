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
