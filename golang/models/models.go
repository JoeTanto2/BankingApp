package models


type User struct {
	Name     string `json:"name,omitempty" validate:"required"`
	Address  string `json:"address,omitempty" validate:"required"`
	Email    string `json:"email,omitempty" validate:"required"`
	Password string `json:"password,omitempty" validate:"required"`
	Age      int    `json:"age,omitempty" validate:"required"`
	Country  string `json:"country,omitempty" validate:"required"`
	State    string `json:"state,omitempty" validate:"required"`
}

type Card struct {
	User               string
	Number             string
	ExpirationDate     string
	SecurityCode       string
	Network            string
	Balance            float32
	TransactionHistory []TransactionHistory
}

type TransactionHistory struct {
	TransferredFrom   string  `bson:"transferredfrom,omitempty"`
	TransferredTo     string  `bson:"transferredto,omitempty"`
	AmountTransferred float32 `bson:"amounttransferred,omitempty"`
	ActionDate        string  `bson:"actiondate,omitempty"`
	Action            string  `action:"action,omitempty"`
}
