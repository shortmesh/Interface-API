package masclient

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

type CreateUserResponse struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Username string `json:"username"`
		} `json:"attributes"`
	} `json:"data"`
}

type CreatePersonalSessionResponse struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			AccessToken string `json:"access_token"`
			ExpiresAt   string `json:"expires_at"`
		} `json:"attributes"`
	} `json:"data"`
}
