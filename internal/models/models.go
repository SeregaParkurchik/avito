package models

type Employee struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Coins    int    `json:"coins"`
	Token    string `json:"token"`
}

type SendCoin struct {
	FromUser string `json:"fromUser"`
	ToUser   string `json:"toUser"`
	Amount   int    `json:"amount"`
}

type ReceivedTransaction struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

type SentTransaction struct {
	FromUser string `json:"fromUser"`
	Amount   int    `json:"amount"`
}

type Item struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type CoinHistory struct {
	Received []ReceivedTransaction `json:"received"`
	Sent     []SentTransaction     `json:"sent"`
}

type InfoResponse struct {
	Coins       int         `json:"coins"`
	Inventory   []Item      `json:"inventory"`
	CoinHistory CoinHistory `json:"coinHistory"`
}
