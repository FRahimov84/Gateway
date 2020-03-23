package app

type Payload struct {
	Id    int64    `json:"id"`
	Exp   int64    `json:"exp"`
	Roles []string `json:"roles"`
}
