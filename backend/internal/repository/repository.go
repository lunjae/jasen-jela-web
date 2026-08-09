package repository

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/lunjae/jasen-jela-web/backend/internal/models"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrCategoryInUse   = errors.New("category in use")
	ErrInvalidCategory = errors.New("invalid category")
	ErrSlugInUse       = errors.New("slug in use")
)

type Store interface {
	ListProducts(context.Context, models.ProductFilter) (models.Page[models.Product], error)
	ProductBySlug(context.Context, string) (models.Product, error)
	ProductByID(context.Context, string) (models.Product, error)
	SaveProduct(context.Context, models.Product) (models.Product, error)
	DeleteProduct(context.Context, string) error
	ListCategories(context.Context, bool) ([]models.Category, error)
	CategoryByID(context.Context, string) (models.Category, error)
	SaveCategory(context.Context, models.Category) (models.Category, error)
	DeleteCategory(context.Context, string) error
	SaveInquiry(context.Context, models.Inquiry) (models.Inquiry, error)
	ListInquiries(context.Context, models.InquiryFilter) (models.Page[models.Inquiry], error)
	InquiryByID(context.Context, string) (models.Inquiry, error)
	UpdateInquiryStatus(context.Context, string, string) (models.Inquiry, error)
	IsAdmin(context.Context, string) (bool, error)
}
type Memory struct {
	mu         sync.RWMutex
	products   map[string]models.Product
	categories map[string]models.Category
	inquiries  map[string]models.Inquiry
	admins     map[string]bool
	next       int
}

func NewMemory() *Memory {
	return &Memory{products: map[string]models.Product{}, categories: map[string]models.Category{}, inquiries: map[string]models.Inquiry{}, admins: map[string]bool{"local-admin": true}, next: 100}
}
func (m *Memory) id(prefix string) string {
	m.next++
	return prefix + time.Now().UTC().Format("20060102") + string(rune(m.next))
}
func (m *Memory) ListProducts(_ context.Context, f models.ProductFilter) (models.Page[models.Product], error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a := []models.Product{}
	for _, p := range m.products {
		if f.PublishedOnly && !p.Published {
			continue
		}
		if f.Search != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(f.Search)) {
			continue
		}
		if f.Category != "" && p.CategoryID != f.Category {
			continue
		}
		if f.Material != "" && !strings.EqualFold(p.Material, f.Material) {
			continue
		}
		if f.Color != "" && !strings.EqualFold(p.Color, f.Color) {
			continue
		}
		if f.MinPrice != nil && (p.Price == nil || *p.Price < *f.MinPrice) {
			continue
		}
		if f.MaxPrice != nil && (p.Price == nil || *p.Price > *f.MaxPrice) {
			continue
		}
		a = append(a, p)
	}
	sort.Slice(a, func(i, j int) bool {
		switch f.Sort {
		case "name_desc":
			return a[i].Name > a[j].Name
		case "price_asc":
			return price(a[i]) < price(a[j])
		case "price_desc":
			return price(a[i]) > price(a[j])
		default:
			return a[i].CreatedAt.After(a[j].CreatedAt)
		}
	})
	return paginate(a, f.Page, f.PageSize), nil
}
func price(p models.Product) float64 {
	if p.Price == nil {
		return 1e18
	}
	return *p.Price
}
func paginate[T any](a []T, page, size int) models.Page[T] {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 12
	}
	if size > 100 {
		size = 100
	}
	start := (page - 1) * size
	if start > len(a) {
		start = len(a)
	}
	end := start + size
	if end > len(a) {
		end = len(a)
	}
	pages := (len(a) + size - 1) / size
	return models.Page[T]{Items: a[start:end], Page: page, PageSize: size, Total: len(a), TotalPages: pages}
}
func (m *Memory) ProductBySlug(_ context.Context, s string) (models.Product, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.products {
		if p.Slug == s {
			return p, nil
		}
	}
	return models.Product{}, ErrNotFound
}
func (m *Memory) ProductByID(_ context.Context, id string) (models.Product, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.products[id]
	if !ok {
		return p, ErrNotFound
	}
	return p, nil
}
func (m *Memory) SaveProduct(_ context.Context, p models.Product) (models.Product, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.categories[p.CategoryID]; !ok {
		return models.Product{}, ErrInvalidCategory
	}
	for id, existing := range m.products {
		if id != p.ID && existing.Slug == p.Slug {
			return models.Product{}, ErrSlugInUse
		}
	}
	now := time.Now().UTC()
	if p.ID == "" {
		p.ID = m.id("p")
		p.CreatedAt = now
	}
	p.UpdatedAt = now
	p.Images = models.NormalizeImages(p.Images)
	m.products[p.ID] = p
	return p, nil
}
func (m *Memory) DeleteProduct(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.products[id]; !ok {
		return ErrNotFound
	}
	delete(m.products, id)
	return nil
}
func (m *Memory) ListCategories(_ context.Context, pub bool) ([]models.Category, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a := []models.Category{}
	for _, c := range m.categories {
		if !pub || c.Published {
			a = append(a, c)
		}
	}
	sort.Slice(a, func(i, j int) bool { return a[i].Name < a[j].Name })
	return a, nil
}
func (m *Memory) CategoryByID(_ context.Context, id string) (models.Category, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.categories[id]
	if !ok {
		return c, ErrNotFound
	}
	return c, nil
}
func (m *Memory) SaveCategory(_ context.Context, c models.Category) (models.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, existing := range m.categories {
		if id != c.ID && existing.Slug == c.Slug {
			return models.Category{}, ErrSlugInUse
		}
	}
	now := time.Now().UTC()
	if c.ID == "" {
		c.ID = m.id("c")
		c.CreatedAt = now
	}
	c.UpdatedAt = now
	m.categories[c.ID] = c
	return c, nil
}
func (m *Memory) DeleteCategory(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range m.products {
		if p.CategoryID == id {
			return ErrCategoryInUse
		}
	}
	if _, ok := m.categories[id]; !ok {
		return ErrNotFound
	}
	delete(m.categories, id)
	return nil
}
func (m *Memory) SaveInquiry(_ context.Context, i models.Inquiry) (models.Inquiry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().UTC()
	i.ID = m.id("i")
	i.Status = "new"
	i.CreatedAt = now
	i.UpdatedAt = now
	m.inquiries[i.ID] = i
	return i, nil
}
func (m *Memory) ListInquiries(_ context.Context, f models.InquiryFilter) (models.Page[models.Inquiry], error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a := []models.Inquiry{}
	for _, i := range m.inquiries {
		if f.Status == "" || i.Status == f.Status {
			a = append(a, i)
		}
	}
	sort.Slice(a, func(x, y int) bool { return a[x].CreatedAt.After(a[y].CreatedAt) })
	return paginate(a, f.Page, f.PageSize), nil
}
func (m *Memory) InquiryByID(_ context.Context, id string) (models.Inquiry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	i, ok := m.inquiries[id]
	if !ok {
		return i, ErrNotFound
	}
	return i, nil
}
func (m *Memory) UpdateInquiryStatus(_ context.Context, id, s string) (models.Inquiry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, ok := m.inquiries[id]
	if !ok {
		return i, ErrNotFound
	}
	i.Status = s
	i.UpdatedAt = time.Now().UTC()
	m.inquiries[id] = i
	return i, nil
}
func (m *Memory) IsAdmin(_ context.Context, id string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.admins[id], nil
}
