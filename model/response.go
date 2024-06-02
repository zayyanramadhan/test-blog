package model

type Response struct {
	Message string `json:"Message"`
}

type ResponseSuccessData struct {
	Message string      `json:"Message"`
	Data    interface{} `json:"Data"`
}
