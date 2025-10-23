package gen

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestRandomOrder_GeneratesUniqueOrders(t *testing.T) {
	gofakeit.Seed(0)

	o1 := RandomOrder()
	o2 := RandomOrder()

	if o1.OrderUID == o2.OrderUID {
		t.Fatalf("expected different OrderUIDs, got same: %s", o1.OrderUID)
	}
	if len(o1.Items) == 0 {
		t.Fatalf("expected order to have items, got 0")
	}
	if o1.Delivery.Name == "" {
		t.Fatalf("expected delivery name to be non-empty")
	}
}

func TestRandomOrder_FieldsFilled(t *testing.T) {
	o := RandomOrder()
	if o.TrackNumber == "" || o.Entry == "" {
		t.Fatalf("expected non-empty TrackNumber and Entry")
	}
	if o.Payment.Amount <= 0 {
		t.Fatalf("expected payment amount > 0")
	}
}
