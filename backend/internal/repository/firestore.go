package repository

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/lunjae/jasen-jela-web/backend/internal/models"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sort"
	"strings"
	"time"
)

type Firestore struct{ client *firestore.Client }

func NewFirestore(c *firestore.Client) *Firestore { return &Firestore{client: c} }
func (f *Firestore) ListProducts(ctx context.Context, q models.ProductFilter) (models.Page[models.Product], error) {
	it := f.client.Collection("products").Documents(ctx)
	defer it.Stop()
	a := []models.Product{}
	for {
		d, e := it.Next()
		if e == iterator.Done {
			break
		}
		if e != nil {
			return models.Page[models.Product]{}, e
		}
		var p models.Product
		if e = d.DataTo(&p); e != nil {
			return models.Page[models.Product]{}, e
		}
		p.ID = d.Ref.ID
		p.Images = models.NormalizeImages(p.Images)
		if q.PublishedOnly && !p.Published {
			continue
		}
		if q.Search != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(q.Search)) {
			continue
		}
		if q.Category != "" && p.CategoryID != q.Category {
			continue
		}
		if q.Material != "" && !strings.EqualFold(p.Material, q.Material) {
			continue
		}
		if q.Color != "" && !strings.EqualFold(p.Color, q.Color) {
			continue
		}
		if q.MinPrice != nil && (p.Price == nil || *p.Price < *q.MinPrice) {
			continue
		}
		if q.MaxPrice != nil && (p.Price == nil || *p.Price > *q.MaxPrice) {
			continue
		}
		a = append(a, p)
	}
	sort.Slice(a, func(i, j int) bool {
		switch q.Sort {
		case "name_asc":
			return a[i].Name < a[j].Name
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
	return paginate(a, q.Page, q.PageSize), nil
}
func (f *Firestore) ProductBySlug(ctx context.Context, s string) (models.Product, error) {
	it := f.client.Collection("products").Where("slug", "==", s).Limit(1).Documents(ctx)
	defer it.Stop()
	d, e := it.Next()
	if e == iterator.Done {
		return models.Product{}, ErrNotFound
	}
	if e != nil {
		return models.Product{}, e
	}
	var p models.Product
	e = d.DataTo(&p)
	p.ID = d.Ref.ID
	p.Images = models.NormalizeImages(p.Images)
	return p, e
}
func (f *Firestore) ProductByID(ctx context.Context, id string) (models.Product, error) {
	d, e := f.client.Collection("products").Doc(id).Get(ctx)
	if e != nil {
		if status.Code(e) == codes.NotFound {
			return models.Product{}, ErrNotFound
		}
		return models.Product{}, e
	}
	var p models.Product
	e = d.DataTo(&p)
	p.ID = id
	p.Images = models.NormalizeImages(p.Images)
	return p, e
}
func (f *Firestore) SaveProduct(ctx context.Context, p models.Product) (models.Product, error) {
	ref := f.client.Collection("products").Doc(p.ID)
	if p.ID == "" {
		ref = f.client.Collection("products").NewDoc()
		p.ID = ref.ID
	}
	now := time.Now().UTC()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now
	p.Images = models.NormalizeImages(p.Images)
	err := f.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		if _, err := tx.Get(f.client.Collection("categories").Doc(p.CategoryID)); err != nil {
			if status.Code(err) == codes.NotFound {
				return ErrInvalidCategory
			}
			return err
		}
		duplicates := tx.Documents(f.client.Collection("products").Where("slug", "==", p.Slug).Limit(2))
		defer duplicates.Stop()
		for {
			doc, err := duplicates.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}
			if doc.Ref.ID != p.ID {
				return ErrSlugInUse
			}
		}
		return tx.Set(ref, p)
	})
	return p, err
}
func (f *Firestore) DeleteProduct(ctx context.Context, id string) error {
	if _, e := f.client.Collection("products").Doc(id).Get(ctx); e != nil {
		if status.Code(e) == codes.NotFound {
			return ErrNotFound
		}
		return e
	}
	_, e := f.client.Collection("products").Doc(id).Delete(ctx)
	return e
}
func (f *Firestore) ListCategories(ctx context.Context, pub bool) ([]models.Category, error) {
	it := f.client.Collection("categories").Documents(ctx)
	defer it.Stop()
	a := []models.Category{}
	for {
		d, e := it.Next()
		if e == iterator.Done {
			break
		}
		if e != nil {
			return nil, e
		}
		var c models.Category
		if e = d.DataTo(&c); e != nil {
			return nil, e
		}
		c.ID = d.Ref.ID
		if !pub || c.Published {
			a = append(a, c)
		}
	}
	sort.Slice(a, func(i, j int) bool { return a[i].Name < a[j].Name })
	return a, nil
}
func (f *Firestore) CategoryByID(ctx context.Context, id string) (models.Category, error) {
	d, e := f.client.Collection("categories").Doc(id).Get(ctx)
	if e != nil {
		if status.Code(e) == codes.NotFound {
			return models.Category{}, ErrNotFound
		}
		return models.Category{}, e
	}
	var c models.Category
	e = d.DataTo(&c)
	c.ID = id
	return c, e
}
func (f *Firestore) SaveCategory(ctx context.Context, c models.Category) (models.Category, error) {
	ref := f.client.Collection("categories").Doc(c.ID)
	if c.ID == "" {
		ref = f.client.Collection("categories").NewDoc()
		c.ID = ref.ID
	}
	now := time.Now().UTC()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	c.UpdatedAt = now
	err := f.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		duplicates := tx.Documents(f.client.Collection("categories").Where("slug", "==", c.Slug).Limit(2))
		defer duplicates.Stop()
		for {
			doc, err := duplicates.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}
			if doc.Ref.ID != c.ID {
				return ErrSlugInUse
			}
		}
		return tx.Set(ref, c)
	})
	return c, err
}
func (f *Firestore) DeleteCategory(ctx context.Context, id string) error {
	ref := f.client.Collection("categories").Doc(id)
	return f.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		if _, err := tx.Get(ref); err != nil {
			if status.Code(err) == codes.NotFound {
				return ErrNotFound
			}
			return err
		}
		used := tx.Documents(f.client.Collection("products").Where("categoryId", "==", id).Limit(1))
		defer used.Stop()
		if _, err := used.Next(); err == nil {
			return ErrCategoryInUse
		} else if err != iterator.Done {
			return err
		}
		return tx.Delete(ref)
	})
}
func (f *Firestore) SaveInquiry(ctx context.Context, i models.Inquiry) (models.Inquiry, error) {
	ref := f.client.Collection("inquiries").NewDoc()
	i.ID = ref.ID
	now := time.Now().UTC()
	i.Status = "new"
	i.CreatedAt = now
	i.UpdatedAt = now
	_, e := ref.Set(ctx, i)
	return i, e
}
func (f *Firestore) ListInquiries(ctx context.Context, q models.InquiryFilter) (models.Page[models.Inquiry], error) {
	it := f.client.Collection("inquiries").Documents(ctx)
	defer it.Stop()
	a := []models.Inquiry{}
	for {
		d, e := it.Next()
		if e == iterator.Done {
			break
		}
		if e != nil {
			return models.Page[models.Inquiry]{}, e
		}
		var i models.Inquiry
		if e = d.DataTo(&i); e != nil {
			return models.Page[models.Inquiry]{}, e
		}
		i.ID = d.Ref.ID
		if q.Status == "" || i.Status == q.Status {
			a = append(a, i)
		}
	}
	sort.Slice(a, func(i, j int) bool { return a[i].CreatedAt.After(a[j].CreatedAt) })
	return paginate(a, q.Page, q.PageSize), nil
}
func (f *Firestore) InquiryByID(ctx context.Context, id string) (models.Inquiry, error) {
	d, e := f.client.Collection("inquiries").Doc(id).Get(ctx)
	if e != nil {
		if status.Code(e) == codes.NotFound {
			return models.Inquiry{}, ErrNotFound
		}
		return models.Inquiry{}, e
	}
	var i models.Inquiry
	e = d.DataTo(&i)
	i.ID = id
	return i, e
}
func (f *Firestore) UpdateInquiryStatus(ctx context.Context, id, s string) (models.Inquiry, error) {
	_, e := f.client.Collection("inquiries").Doc(id).Update(ctx, []firestore.Update{{Path: "status", Value: s}, {Path: "updatedAt", Value: firestore.ServerTimestamp}})
	if e != nil {
		if status.Code(e) == codes.NotFound {
			return models.Inquiry{}, ErrNotFound
		}
		return models.Inquiry{}, e
	}
	return f.InquiryByID(ctx, id)
}
func (f *Firestore) IsAdmin(ctx context.Context, uid string) (bool, error) {
	d, e := f.client.Collection("admins").Doc(uid).Get(ctx)
	if e != nil {
		if status.Code(e) == codes.NotFound {
			return false, nil
		}
		return false, e
	}
	return adminEnabled(d.Data()), nil
}

func adminEnabled(data map[string]any) bool {
	enabled, ok := data["enabled"].(bool)
	return ok && enabled
}
