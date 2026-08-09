package models

import (
	"sort"
	"time"
)

type ProductImage struct {
	URL               string `json:"url" firestore:"url"`
	PublicID          string `json:"publicId" firestore:"publicId"`
	LegacyStoragePath string `json:"-" firestore:"storagePath,omitempty"`
	Alt               string `json:"alt" firestore:"alt"`
	Order             int    `json:"order" firestore:"order"`
	IsPrimary         bool   `json:"isPrimary" firestore:"isPrimary"`
}

// NormalizeImages migrates legacy storagePath values in memory, preserves the
// requested primary image, and guarantees stable contiguous ordering.
func NormalizeImages(images []ProductImage) []ProductImage {
	if images == nil {
		return []ProductImage{}
	}
	for i := range images {
		if images[i].PublicID == "" {
			images[i].PublicID = images[i].LegacyStoragePath
		}
		images[i].LegacyStoragePath = ""
	}
	sort.SliceStable(images, func(i, j int) bool { return images[i].Order < images[j].Order })
	primary := -1
	for i := range images {
		images[i].Order = i
		if images[i].IsPrimary && primary == -1 {
			primary = i
		} else {
			images[i].IsPrimary = false
		}
	}
	if len(images) > 0 && primary == -1 {
		images[0].IsPrimary = true
	}
	return images
}

type Dimensions struct {
	Length *float64 `json:"length,omitempty" firestore:"length,omitempty"`
	Width  *float64 `json:"width,omitempty" firestore:"width,omitempty"`
	Height *float64 `json:"height,omitempty" firestore:"height,omitempty"`
}
type Product struct {
	ID               string         `json:"id" firestore:"-"`
	Name             string         `json:"name" firestore:"name"`
	Slug             string         `json:"slug" firestore:"slug"`
	ShortDescription string         `json:"shortDescription" firestore:"shortDescription"`
	Description      string         `json:"description" firestore:"description"`
	CategoryID       string         `json:"categoryId" firestore:"categoryId"`
	Material         string         `json:"material" firestore:"material"`
	Color            string         `json:"color" firestore:"color"`
	Dimensions       *Dimensions    `json:"dimensions,omitempty" firestore:"dimensions,omitempty"`
	Price            *float64       `json:"price,omitempty" firestore:"price,omitempty"`
	Currency         string         `json:"currency" firestore:"currency"`
	Images           []ProductImage `json:"images" firestore:"images"`
	Featured         bool           `json:"featured" firestore:"featured"`
	Published        bool           `json:"published" firestore:"published"`
	CreatedAt        time.Time      `json:"createdAt" firestore:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt" firestore:"updatedAt"`
}
type Category struct {
	ID          string    `json:"id" firestore:"-"`
	Name        string    `json:"name" firestore:"name"`
	Slug        string    `json:"slug" firestore:"slug"`
	Description string    `json:"description,omitempty" firestore:"description,omitempty"`
	Published   bool      `json:"published" firestore:"published"`
	CreatedAt   time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" firestore:"updatedAt"`
}
type Inquiry struct {
	ID        string    `json:"id" firestore:"-"`
	ProductID string    `json:"productId,omitempty" firestore:"productId,omitempty"`
	FullName  string    `json:"fullName" firestore:"fullName"`
	Email     string    `json:"email" firestore:"email"`
	Phone     string    `json:"phone" firestore:"phone"`
	Message   string    `json:"message" firestore:"message"`
	Status    string    `json:"status" firestore:"status"`
	CreatedAt time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" firestore:"updatedAt"`
}
type ProductFilter struct {
	Search, Category, Material, Color, Sort string
	MinPrice, MaxPrice                      *float64
	PublishedOnly                           bool
	Page, PageSize                          int
}
type InquiryFilter struct {
	Status         string
	Page, PageSize int
}
type Page[T any] struct {
	Items      []T `json:"items"`
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}
