package validation

import (
	"github.com/lunjae/jasen-jela-web/backend/internal/models"
	"testing"
)

func TestValidInquiry(t *testing.T) {
	got := Inquiry(models.Inquiry{FullName: "Petar Petrović", Email: "petar@example.com", Phone: "+381 64 123 456", Message: "Molim vas za informacije o dostupnosti."})
	if len(got) > 0 {
		t.Fatalf("unexpected errors: %#v", got)
	}
}
func TestProductRejectsCurrencyAndNegativePrice(t *testing.T) {
	n := -1.0
	got := Product(models.Product{Name: "Model A", Slug: "model-a", ShortDescription: "Dovoljno dug kratak opis", Description: "Detaljan opis proizvoda koji je dovoljno dug.", CategoryID: "cat", Material: "Hrast", Color: "Orah", Currency: "USD", Price: &n})
	if got["currency"] == "" || got["price"] == "" {
		t.Fatalf("missing errors: %#v", got)
	}
}

func TestProductRejectsInvalidCloudinaryMetadata(t *testing.T) {
	got := Product(models.Product{Name: "Model A", Slug: "model-a", ShortDescription: "Dovoljno dug kratak opis", Description: "Detaljan opis proizvoda koji je dovoljno dug.", CategoryID: "cat", Material: "Hrast", Color: "Orah", Currency: "EUR", Images: []models.ProductImage{{URL: "data:image/png;base64,fake", PublicID: "wrong/path"}}})
	if got["images[0]"] == "" {
		t.Fatalf("missing image metadata error: %#v", got)
	}
}

func TestCategoryRejectsInvalidSlug(t *testing.T) {
	got := Category(models.Category{Name: "Valid name", Slug: "Not Valid"})
	if got["slug"] == "" {
		t.Fatalf("missing slug error: %#v", got)
	}
}
