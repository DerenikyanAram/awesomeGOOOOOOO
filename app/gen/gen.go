package gen

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

func init() { gofakeit.Seed(0) }

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ChrtID      int64  `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	RID         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int64  `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

type Order struct {
	OrderUID          string    `json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Delivery          Delivery  `json:"delivery"`
	Payment           Payment   `json:"payment"`
	Items             []Item    `json:"items"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	Shardkey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
}

func RandomOrder() Order {
	n := gofakeit.Numerify("#######")
	items := make([]Item, 0, 3)
	for i := 0; i < 3; i++ {
		items = append(items, Item{
			ChrtID:      gofakeit.Int64(),
			TrackNumber: "TR" + n,
			Price:       gofakeit.Number(100, 100000),
			RID:         gofakeit.UUID(),
			Name:        gofakeit.ProductName(),
			Sale:        gofakeit.Number(0, 70),
			Size:        gofakeit.RandomString([]string{"XS", "S", "M", "L", "XL"}),
			TotalPrice:  gofakeit.Number(100, 100000),
			NmID:        gofakeit.Int64(),
			Brand:       gofakeit.Company(),
			Status:      gofakeit.Number(1, 10),
		})
	}
	return Order{
		OrderUID:          "ORD-" + n,
		TrackNumber:       "TR" + n,
		Entry:             "WBIL",
		Delivery:          Delivery{Name: gofakeit.Name(), Phone: gofakeit.Phone(), Zip: gofakeit.Zip(), City: gofakeit.City(), Address: gofakeit.Street(), Region: gofakeit.State(), Email: gofakeit.Email()},
		Payment:           Payment{Transaction: gofakeit.UUID(), RequestID: "", Currency: "RUB", Provider: "wbpay", Amount: gofakeit.Number(1000, 500000), PaymentDt: time.Now().Unix(), Bank: gofakeit.Company(), DeliveryCost: gofakeit.Number(0, 10000), GoodsTotal: gofakeit.Number(1000, 500000), CustomFee: 0},
		Items:             items,
		Locale:            "ru",
		InternalSignature: "",
		CustomerID:        gofakeit.Username(),
		DeliveryService:   "meest",
		Shardkey:          gofakeit.Numerify("##"),
		SmID:              gofakeit.Number(1, 999),
		DateCreated:       time.Now().UTC(),
		OofShard:          gofakeit.Numerify("##"),
	}
}
