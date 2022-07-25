package model

type Account struct {
	ID   string      `json:"id"`
	Name string      `json:"name"`
	Type AccountType `json:"type"`
}

type AccountFilter struct {
	Name string `json:"name,omitempty"`
	//Types []AccountType `json:"types,omitempty"`
}
