package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/lunjae/jasen-jela-web/backend/internal/models"
)

func TestMemoryRejectsProductWithMissingCategory(t *testing.T) {
	store := NewMemory()
	_, err := store.SaveProduct(context.Background(), models.Product{Slug: "model"})
	if !errors.Is(err, ErrInvalidCategory) {
		t.Fatalf("expected ErrInvalidCategory, got %v", err)
	}
}

func TestMemoryPreventsDeletingUsedCategory(t *testing.T) {
	store := NewMemory()
	category, err := store.SaveCategory(context.Background(), models.Category{Name: "Category", Slug: "category"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.SaveProduct(context.Background(), models.Product{CategoryID: category.ID, Slug: "model"}); err != nil {
		t.Fatal(err)
	}
	if err = store.DeleteCategory(context.Background(), category.ID); !errors.Is(err, ErrCategoryInUse) {
		t.Fatalf("expected ErrCategoryInUse, got %v", err)
	}
}

func TestAdminEnabledRequiresExplicitTrue(t *testing.T) {
	if adminEnabled(map[string]any{}) || adminEnabled(map[string]any{"enabled": false}) || !adminEnabled(map[string]any{"enabled": true}) {
		t.Fatal("admin enabled check is not strict")
	}
}
