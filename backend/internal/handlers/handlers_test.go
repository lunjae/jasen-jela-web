package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/lunjae/jasen-jela-web/backend/internal/models"
	"github.com/lunjae/jasen-jela-web/backend/internal/repository"
	"net/http"
	"net/http/httptest"
	"testing"
)

type imageUploaderStub struct {
	uploadImages []models.ProductImage
	uploadErr    error
	deleteErr    error
	deleteErrors map[string]error
	deleted      []string
}

func (u *imageUploaderStub) Upload(*http.Request, string) ([]models.ProductImage, error) {
	return u.uploadImages, u.uploadErr
}
func (u *imageUploaderStub) Delete(_ *http.Request, publicID string) error {
	u.deleted = append(u.deleted, publicID)
	if err := u.deleteErrors[publicID]; err != nil {
		return err
	}
	return u.deleteErr
}

func TestHealth(t *testing.T) {
	h := New(repository.NewMemory(), nil)
	r := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	h.Health(w, r)
	if w.Code != 200 {
		t.Fatalf("status=%d", w.Code)
	}
	var body struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if e := json.NewDecoder(w.Body).Decode(&body); e != nil || body.Data.Status != "ok" {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}
func TestInquiryValidation(t *testing.T) {
	h := New(repository.NewMemory(), nil)
	r := httptest.NewRequest(http.MethodPost, "/api/inquiries", bytes.NewBufferString(`{"fullName":"A","email":"bad","phone":"x","message":"short"}`))
	w := httptest.NewRecorder()
	h.CreateInquiry(w, r)
	if w.Code != 400 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("validation_error")) {
		t.Fatal("missing validation code")
	}
}

func TestSetPrimaryImageKeepsExactlyOnePrimary(t *testing.T) {
	store := repository.NewMemory()
	category, _ := store.SaveCategory(t.Context(), models.Category{Name: "Kategorija", Slug: "kategorija"})
	product, _ := store.SaveProduct(t.Context(), models.Product{CategoryID: category.ID, Slug: "model", Images: []models.ProductImage{
		{URL: "https://example.com/one.jpg", PublicID: "jasen-jela/products/p/one", Order: 0, IsPrimary: true},
		{URL: "https://example.com/two.jpg", PublicID: "jasen-jela/products/p/two", Order: 1},
	}})
	h := New(store, &imageUploaderStub{})
	r := httptest.NewRequest(http.MethodPatch, "/api/admin/products/"+product.ID+"/images/primary", bytes.NewBufferString(`{"publicId":"jasen-jela/products/p/two"}`))
	r.SetPathValue("id", product.ID)
	w := httptest.NewRecorder()
	h.SetPrimaryImage(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	saved, _ := store.ProductByID(t.Context(), product.ID)
	if saved.Images[0].IsPrimary || !saved.Images[1].IsPrimary {
		t.Fatalf("unexpected primary flags: %#v", saved.Images)
	}
}

func TestDeleteImageKeepsMetadataWhenCloudinaryFails(t *testing.T) {
	store := repository.NewMemory()
	category, _ := store.SaveCategory(t.Context(), models.Category{Name: "Kategorija", Slug: "kategorija"})
	product, _ := store.SaveProduct(t.Context(), models.Product{CategoryID: category.ID, Slug: "model", Images: []models.ProductImage{{URL: "https://example.com/one.jpg", PublicID: "jasen-jela/products/p/one", Order: 0, IsPrimary: true}}})
	h := New(store, &imageUploaderStub{deleteErr: errors.New("cloud unavailable")})
	r := httptest.NewRequest(http.MethodDelete, "/api/admin/products/"+product.ID+"/images", bytes.NewBufferString(`{"publicId":"jasen-jela/products/p/one"}`))
	r.SetPathValue("id", product.ID)
	w := httptest.NewRecorder()
	h.DeleteImage(w, r)
	if w.Code != http.StatusBadGateway {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	saved, _ := store.ProductByID(t.Context(), product.ID)
	if len(saved.Images) != 1 {
		t.Fatalf("image metadata was removed after failed Cloudinary delete: %#v", saved.Images)
	}
}

func TestFirstUploadedImageBecomesPrimary(t *testing.T) {
	store, product := imageTestProduct(t, nil)
	uploader := &imageUploaderStub{uploadImages: []models.ProductImage{{URL: "https://example.com/one.jpg", PublicID: "jasen-jela/products/p/one", Alt: "Prva"}}}
	h := New(store, uploader)
	r := httptest.NewRequest(http.MethodPost, "/api/admin/products/"+product.ID+"/images", nil)
	r.SetPathValue("id", product.ID)
	w := httptest.NewRecorder()
	h.UploadImage(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	saved, _ := store.ProductByID(t.Context(), product.ID)
	if len(saved.Images) != 1 || !saved.Images[0].IsPrimary || saved.Images[0].Order != 0 {
		t.Fatalf("unexpected images: %#v", saved.Images)
	}
}

func TestUploadedPrimaryClearsPreviousPrimary(t *testing.T) {
	existing := []models.ProductImage{{URL: "https://example.com/one.jpg", PublicID: "jasen-jela/products/p/one", Order: 0, IsPrimary: true}}
	store, product := imageTestProduct(t, existing)
	uploader := &imageUploaderStub{uploadImages: []models.ProductImage{{URL: "https://example.com/two.jpg", PublicID: "jasen-jela/products/p/two", Alt: "Druga", IsPrimary: true}}}
	h := New(store, uploader)
	r := httptest.NewRequest(http.MethodPost, "/api/admin/products/"+product.ID+"/images", nil)
	r.SetPathValue("id", product.ID)
	w := httptest.NewRecorder()
	h.UploadImage(w, r)
	saved, _ := store.ProductByID(t.Context(), product.ID)
	if w.Code != http.StatusCreated || saved.Images[0].IsPrimary || !saved.Images[1].IsPrimary {
		t.Fatalf("status=%d images=%#v", w.Code, saved.Images)
	}
}

func TestDeletePrimaryRemovesMetadataAndPromotesFirstRemaining(t *testing.T) {
	images := []models.ProductImage{
		{URL: "https://example.com/one.jpg", PublicID: "jasen-jela/products/p/one", Order: 0, IsPrimary: true},
		{URL: "https://example.com/two.jpg", PublicID: "jasen-jela/products/p/two", Order: 1},
	}
	store, product := imageTestProduct(t, images)
	uploader := &imageUploaderStub{}
	h := New(store, uploader)
	r := httptest.NewRequest(http.MethodDelete, "/api/admin/products/"+product.ID+"/images", bytes.NewBufferString(`{"publicId":"jasen-jela/products/p/one"}`))
	r.SetPathValue("id", product.ID)
	w := httptest.NewRecorder()
	h.DeleteImage(w, r)
	saved, _ := store.ProductByID(t.Context(), product.ID)
	if w.Code != http.StatusOK || len(saved.Images) != 1 || saved.Images[0].PublicID != "jasen-jela/products/p/two" || !saved.Images[0].IsPrimary || saved.Images[0].Order != 0 {
		t.Fatalf("status=%d images=%#v", w.Code, saved.Images)
	}
	if len(uploader.deleted) != 1 || uploader.deleted[0] != "jasen-jela/products/p/one" {
		t.Fatalf("unexpected Cloudinary deletes: %#v", uploader.deleted)
	}
}

func TestInvalidImagePublicIDReturnsNotFound(t *testing.T) {
	store, product := imageTestProduct(t, []models.ProductImage{{URL: "https://example.com/one.jpg", PublicID: "jasen-jela/products/p/one", Order: 0, IsPrimary: true}})
	h := New(store, &imageUploaderStub{})
	r := httptest.NewRequest(http.MethodDelete, "/api/admin/products/"+product.ID+"/images", bytes.NewBufferString(`{"publicId":"jasen-jela/products/p/missing"}`))
	r.SetPathValue("id", product.ID)
	w := httptest.NewRecorder()
	h.DeleteImage(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestReorderImagesRequiresExactImageSet(t *testing.T) {
	images := []models.ProductImage{
		{URL: "https://example.com/one.jpg", PublicID: "jasen-jela/products/p/one", Order: 0, IsPrimary: true},
		{URL: "https://example.com/two.jpg", PublicID: "jasen-jela/products/p/two", Order: 1},
	}
	store, product := imageTestProduct(t, images)
	h := New(store, &imageUploaderStub{})
	r := httptest.NewRequest(http.MethodPatch, "/api/admin/products/"+product.ID+"/images/order", bytes.NewBufferString(`{"publicIds":["jasen-jela/products/p/two","jasen-jela/products/p/two"]}`))
	r.SetPathValue("id", product.ID)
	w := httptest.NewRecorder()
	h.ReorderImages(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	r = httptest.NewRequest(http.MethodPatch, "/api/admin/products/"+product.ID+"/images/order", bytes.NewBufferString(`{"publicIds":["jasen-jela/products/p/two","jasen-jela/products/p/one"]}`))
	r.SetPathValue("id", product.ID)
	w = httptest.NewRecorder()
	h.ReorderImages(w, r)
	saved, _ := store.ProductByID(t.Context(), product.ID)
	if w.Code != http.StatusOK || saved.Images[0].PublicID != "jasen-jela/products/p/two" || saved.Images[0].Order != 0 || !saved.Images[1].IsPrimary {
		t.Fatalf("status=%d images=%#v", w.Code, saved.Images)
	}
}

func TestDeleteProductReportsPartialCloudinaryCleanup(t *testing.T) {
	images := []models.ProductImage{
		{URL: "https://example.com/one.jpg", PublicID: "jasen-jela/products/p/one", Order: 0, IsPrimary: true},
		{URL: "https://example.com/two.jpg", PublicID: "jasen-jela/products/p/two", Order: 1},
	}
	store, product := imageTestProduct(t, images)
	uploader := &imageUploaderStub{deleteErrors: map[string]error{"jasen-jela/products/p/two": errors.New("cloud unavailable")}}
	h := New(store, uploader)
	r := httptest.NewRequest(http.MethodDelete, "/api/admin/products/"+product.ID, nil)
	r.SetPathValue("id", product.ID)
	w := httptest.NewRecorder()
	h.DeleteProduct(w, r)
	saved, err := store.ProductByID(t.Context(), product.ID)
	if w.Code != http.StatusBadGateway || err != nil || len(saved.Images) != 1 || saved.Images[0].PublicID != "jasen-jela/products/p/two" {
		t.Fatalf("status=%d err=%v images=%#v", w.Code, err, saved.Images)
	}
}

func imageTestProduct(t *testing.T, images []models.ProductImage) (*repository.Memory, models.Product) {
	t.Helper()
	store := repository.NewMemory()
	category, err := store.SaveCategory(t.Context(), models.Category{Name: "Kategorija", Slug: "kategorija"})
	if err != nil {
		t.Fatal(err)
	}
	product, err := store.SaveProduct(t.Context(), models.Product{CategoryID: category.ID, Slug: "model", Name: "Model", Images: images})
	if err != nil {
		t.Fatal(err)
	}
	return store, product
}
