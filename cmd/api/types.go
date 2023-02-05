package main

type InputAttributes[T ComplexInputModels] struct {
	Id         string `json:"id,omitempty"`
	Type       string `json:"type"`
	Attributes T      `json:"attributes"`
}
type Input[T ComplexInputModels] struct {
	Data InputAttributes[T] `json:"data"`
}

type ComplexInputModels interface {
	TokensAttributes | ItemAttributes | UserAttributes | ActivationAttributes
}

type ItemAttributes struct {
	ListId       *int64   `json:"list_id"`
	Name         *string  `json:"name"`
	Description  *string  `json:"description"`
	Quantity     *int32   `json:"quantity"`
	QuantityType *string  `json:"quantity_type"`
	Price        *float32 `json:"price"`
	IsStarred    *bool    `json:"is_starred"`
	File         *string  `json:"file"`
	Order        *int32   `json:"order"`
}

type UserAttributes struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"Password"`
}

type ActivationAttributes struct {
	Token string `json:"token"`
}

type TokensAttributes struct {
	Email    string `json:"email"`
	Password string `json:"Password"`
}
