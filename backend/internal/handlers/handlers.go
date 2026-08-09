package handlers

import (
	"encoding/json"
	"errors"
	"github.com/lunjae/jasen-jela-web/backend/internal/models"
	"github.com/lunjae/jasen-jela-web/backend/internal/repository"
	"github.com/lunjae/jasen-jela-web/backend/internal/response"
	"github.com/lunjae/jasen-jela-web/backend/internal/services"
	"github.com/lunjae/jasen-jela-web/backend/internal/validation"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const maxJSON = 1 << 20

type Handler struct {
	store    repository.Store
	uploader ImageUploader
}
type ImageUploader interface {
	Upload(r *http.Request, productID string) ([]models.ProductImage, error)
	Delete(r *http.Request, path string) error
}

func New(s repository.Store, u ImageUploader) *Handler { return &Handler{store: s, uploader: u} }
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	response.Write(w, 200, map[string]string{"status": "ok", "message": "API je dostupna."})
}
func queryInt(r *http.Request, k string, d int) int {
	n, e := strconv.Atoi(r.URL.Query().Get(k))
	if e != nil || n < 1 {
		return d
	}
	return n
}
func queryFloat(r *http.Request, k string) *float64 {
	v := r.URL.Query().Get(k)
	if v == "" {
		return nil
	}
	n, e := strconv.ParseFloat(v, 64)
	if e != nil {
		return nil
	}
	return &n
}
func (h *Handler) Products(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	p, e := h.store.ListProducts(r.Context(), models.ProductFilter{Search: q.Get("search"), Category: q.Get("category"), Material: q.Get("material"), Color: q.Get("color"), Sort: q.Get("sort"), MinPrice: queryFloat(r, "minPrice"), MaxPrice: queryFloat(r, "maxPrice"), PublishedOnly: true, Page: queryInt(r, "page", 1), PageSize: queryInt(r, "pageSize", 12)})
	if e != nil {
		response.WriteError(w, e)
		return
	}
	response.Write(w, 200, p)
}
func (h *Handler) Product(w http.ResponseWriter, r *http.Request) {
	p, e := h.store.ProductBySlug(r.Context(), r.PathValue("slug"))
	if errors.Is(e, repository.ErrNotFound) || (!p.Published && e == nil) {
		response.WriteError(w, response.NotFound("Proizvod nije pronađen."))
		return
	}
	if e != nil {
		response.WriteError(w, e)
		return
	}
	response.Write(w, 200, p)
}
func (h *Handler) Categories(w http.ResponseWriter, r *http.Request) {
	a, e := h.store.ListCategories(r.Context(), true)
	if e != nil {
		response.WriteError(w, e)
		return
	}
	response.Write(w, 200, a)
}
func decode(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSON)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		return response.BadRequest(map[string]string{"body": "Neispravan JSON zahtev."})
	}
	if e := d.Decode(&struct{}{}); e != io.EOF {
		return response.BadRequest(map[string]string{"body": "Zahtev mora sadržati jedan JSON objekat."})
	}
	return nil
}
func (h *Handler) CreateInquiry(w http.ResponseWriter, r *http.Request) {
	var i models.Inquiry
	if e := decode(w, r, &i); e != nil {
		response.WriteError(w, e)
		return
	}
	if f := validation.Inquiry(i); len(f) > 0 {
		response.WriteError(w, response.BadRequest(f))
		return
	}
	if i.ProductID != "" {
		product, e := h.store.ProductByID(r.Context(), i.ProductID)
		if errors.Is(e, repository.ErrNotFound) || (e == nil && !product.Published) {
			response.WriteError(w, response.BadRequest(map[string]string{"productId": "Izabrani proizvod nije dostupan."}))
			return
		}
		if e != nil {
			response.WriteError(w, e)
			return
		}
	}
	now := time.Now().UTC()
	i.Status = "new"
	i.CreatedAt = now
	i.UpdatedAt = now
	saved, e := h.store.SaveInquiry(r.Context(), i)
	if e != nil {
		response.WriteError(w, e)
		return
	}
	response.Write(w, 201, saved)
}
func (h *Handler) AdminProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	p, e := h.store.ListProducts(r.Context(), models.ProductFilter{Search: q.Get("search"), Category: q.Get("category"), Sort: q.Get("sort"), Page: queryInt(r, "page", 1), PageSize: queryInt(r, "pageSize", 20)})
	if e != nil {
		response.WriteError(w, e)
		return
	}
	response.Write(w, 200, p)
}
func (h *Handler) AdminProduct(w http.ResponseWriter, r *http.Request) {
	p, e := h.store.ProductByID(r.Context(), r.PathValue("id"))
	if errors.Is(e, repository.ErrNotFound) {
		response.WriteError(w, response.NotFound("Proizvod nije pronađen."))
		return
	}
	if e != nil {
		response.WriteError(w, e)
		return
	}
	response.Write(w, 200, p)
}
func (h *Handler) SaveProduct(w http.ResponseWriter, r *http.Request) {
	var p models.Product
	if e := decode(w, r, &p); e != nil {
		response.WriteError(w, e)
		return
	}
	if r.Method == http.MethodPut {
		p.ID = r.PathValue("id")
		old, e := h.store.ProductByID(r.Context(), p.ID)
		if e != nil {
			response.WriteError(w, response.NotFound("Proizvod nije pronađen."))
			return
		}
		p.CreatedAt = old.CreatedAt
		p.Images = old.Images
	} else {
		p.ID = ""
		p.CreatedAt = time.Time{}
		p.Images = []models.ProductImage{}
	}
	p.Slug = services.Slug(p.Name)
	if f := validation.Product(p); len(f) > 0 {
		response.WriteError(w, response.BadRequest(f))
		return
	}
	now := time.Now().UTC()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now
	saved, e := h.store.SaveProduct(r.Context(), p)
	if e != nil {
		writeStoreError(w, e)
		return
	}
	status := 200
	if r.Method == http.MethodPost {
		status = 201
	}
	response.Write(w, status, saved)
}
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	p, e := h.store.ProductByID(r.Context(), r.PathValue("id"))
	if errors.Is(e, repository.ErrNotFound) {
		response.WriteError(w, response.NotFound("Proizvod nije pronađen."))
		return
	}
	if e != nil {
		response.WriteError(w, e)
		return
	}
	remaining := make([]models.ProductImage, 0, len(p.Images))
	for _, image := range p.Images {
		if e = h.uploader.Delete(r, image.PublicID); e != nil {
			remaining = append(remaining, image)
		}
	}
	if len(remaining) > 0 {
		p.Images = models.NormalizeImages(remaining)
		if _, saveErr := h.store.SaveProduct(r.Context(), p); saveErr != nil {
			response.WriteError(w, saveErr)
			return
		}
		response.WriteError(w, &response.APIError{Status: 502, Code: "image_cleanup_failed", Message: "Neke slike nisu obrisane iz Cloudinary-ja. Pokušajte ponovo."})
		return
	}
	if len(p.Images) > 0 {
		p.Images = []models.ProductImage{}
		if _, e = h.store.SaveProduct(r.Context(), p); e != nil {
			response.WriteError(w, e)
			return
		}
	}
	e = h.store.DeleteProduct(r.Context(), p.ID)
	if e != nil {
		response.WriteError(w, e)
		return
	}
	response.Write(w, 200, map[string]bool{"deleted": true})
}
func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if h.uploader == nil {
		response.WriteError(w, &response.APIError{Status: 503, Code: "storage_unavailable", Message: "Skladište slika nije konfigurisano."})
		return
	}
	id := r.PathValue("id")
	p, e := h.store.ProductByID(r.Context(), id)
	if e != nil {
		response.WriteError(w, response.NotFound("Proizvod nije pronađen."))
		return
	}
	images, e := h.uploader.Upload(r, id)
	if e != nil {
		if errors.Is(e, services.ErrInvalidImage) {
			message := strings.TrimPrefix(e.Error(), services.ErrInvalidImage.Error()+": ")
			response.WriteError(w, response.BadRequest(map[string]string{"image": message}))
		} else {
			response.WriteError(w, &response.APIError{Status: 502, Code: "image_upload_failed", Message: "Slika trenutno ne može biti sačuvana. Pokušajte ponovo."})
		}
		return
	}
	requestedPrimary := false
	for i := range images {
		if images[i].Alt == "" {
			images[i].Alt = p.Name
		}
		images[i].Order = len(p.Images) + i
		requestedPrimary = requestedPrimary || images[i].IsPrimary
		images[i].IsPrimary = false
	}
	if len(p.Images) == 0 && len(images) > 0 {
		images[0].IsPrimary = true
	} else if requestedPrimary && len(images) > 0 {
		for i := range p.Images {
			p.Images[i].IsPrimary = false
		}
		images[0].IsPrimary = true
	}
	p.Images = append(p.Images, images...)
	p.Images = models.NormalizeImages(p.Images)
	saved, e := h.store.SaveProduct(r.Context(), p)
	if e != nil {
		for _, image := range images {
			_ = h.uploader.Delete(r, image.PublicID)
		}
		response.WriteError(w, e)
		return
	}
	response.Write(w, 201, saved)
}
func (h *Handler) ReorderImages(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PublicIDs []string `json:"publicIds"`
	}
	if e := decode(w, r, &body); e != nil {
		response.WriteError(w, e)
		return
	}
	p, e := h.store.ProductByID(r.Context(), r.PathValue("id"))
	if errors.Is(e, repository.ErrNotFound) {
		response.WriteError(w, response.NotFound("Proizvod nije pronađen."))
		return
	}
	if e != nil {
		response.WriteError(w, e)
		return
	}
	if len(body.PublicIDs) != len(p.Images) {
		response.WriteError(w, response.BadRequest(map[string]string{"publicIds": "Pošaljite sve slike tačno jednom."}))
		return
	}
	byID := make(map[string]models.ProductImage, len(p.Images))
	for _, image := range p.Images {
		byID[image.PublicID] = image
	}
	reordered := make([]models.ProductImage, 0, len(p.Images))
	for order, publicID := range body.PublicIDs {
		image, ok := byID[publicID]
		if !ok {
			response.WriteError(w, response.BadRequest(map[string]string{"publicIds": "Lista slika nije ispravna."}))
			return
		}
		delete(byID, publicID)
		image.Order = order
		reordered = append(reordered, image)
	}
	if len(byID) != 0 {
		response.WriteError(w, response.BadRequest(map[string]string{"publicIds": "Pošaljite sve slike tačno jednom."}))
		return
	}
	p.Images = reordered
	saved, e := h.store.SaveProduct(r.Context(), p)
	if e != nil {
		response.WriteError(w, e)
		return
	}
	response.Write(w, 200, saved)
}
func (h *Handler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PublicID string `json:"publicId"`
	}
	if e := decode(w, r, &body); e != nil {
		response.WriteError(w, e)
		return
	}
	p, e := h.store.ProductByID(r.Context(), r.PathValue("id"))
	if errors.Is(e, repository.ErrNotFound) {
		response.WriteError(w, response.NotFound("Proizvod nije pronađen."))
		return
	}
	if e != nil {
		response.WriteError(w, e)
		return
	}
	found := false
	next := []models.ProductImage{}
	for _, x := range p.Images {
		if x.PublicID == body.PublicID {
			found = true
		} else {
			next = append(next, x)
		}
	}
	if !found {
		response.WriteError(w, response.NotFound("Slika nije pronađena."))
		return
	}
	if e = h.uploader.Delete(r, body.PublicID); e != nil {
		response.WriteError(w, &response.APIError{Status: 502, Code: "cloudinary_delete_failed", Message: "Slika nije obrisana iz Cloudinary-ja. Pokušajte ponovo."})
		return
	}
	p.Images = models.NormalizeImages(next)
	saved, e := h.store.SaveProduct(r.Context(), p)
	if e != nil {
		response.WriteError(w, e)
		return
	}
	response.Write(w, 200, saved)
}
func (h *Handler) SetPrimaryImage(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PublicID string `json:"publicId"`
	}
	if e := decode(w, r, &body); e != nil {
		response.WriteError(w, e)
		return
	}
	p, e := h.store.ProductByID(r.Context(), r.PathValue("id"))
	if errors.Is(e, repository.ErrNotFound) {
		response.WriteError(w, response.NotFound("Proizvod nije pronađen."))
		return
	}
	if e != nil {
		response.WriteError(w, e)
		return
	}
	found := false
	for i := range p.Images {
		p.Images[i].IsPrimary = p.Images[i].PublicID == body.PublicID
		found = found || p.Images[i].IsPrimary
	}
	if !found {
		response.WriteError(w, response.NotFound("Slika nije pronađena."))
		return
	}
	p.Images = models.NormalizeImages(p.Images)
	saved, e := h.store.SaveProduct(r.Context(), p)
	if e != nil {
		response.WriteError(w, e)
		return
	}
	response.Write(w, 200, saved)
}
func (h *Handler) AdminCategories(w http.ResponseWriter, r *http.Request) {
	a, e := h.store.ListCategories(r.Context(), false)
	if e != nil {
		response.WriteError(w, e)
		return
	}
	response.Write(w, 200, a)
}
func (h *Handler) SaveCategory(w http.ResponseWriter, r *http.Request) {
	var c models.Category
	if e := decode(w, r, &c); e != nil {
		response.WriteError(w, e)
		return
	}
	if r.Method == http.MethodPut {
		c.ID = r.PathValue("id")
		old, e := h.store.CategoryByID(r.Context(), c.ID)
		if e != nil {
			response.WriteError(w, response.NotFound("Kategorija nije pronađena."))
			return
		}
		c.CreatedAt = old.CreatedAt
	} else {
		c.ID = ""
		c.CreatedAt = time.Time{}
	}
	c.Slug = services.Slug(c.Name)
	if f := validation.Category(c); len(f) > 0 {
		response.WriteError(w, response.BadRequest(f))
		return
	}
	now := time.Now().UTC()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	c.UpdatedAt = now
	saved, e := h.store.SaveCategory(r.Context(), c)
	if e != nil {
		writeStoreError(w, e)
		return
	}
	status := 200
	if r.Method == http.MethodPost {
		status = 201
	}
	response.Write(w, status, saved)
}

func writeStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrInvalidCategory):
		response.WriteError(w, response.BadRequest(map[string]string{"categoryId": "Izabrana kategorija ne postoji."}))
	case errors.Is(err, repository.ErrSlugInUse):
		response.WriteError(w, &response.APIError{Status: 409, Code: "slug_in_use", Message: "Naziv generiše slug koji se već koristi."})
	default:
		response.WriteError(w, err)
	}
}
func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	e := h.store.DeleteCategory(r.Context(), r.PathValue("id"))
	if errors.Is(e, repository.ErrNotFound) {
		response.WriteError(w, response.NotFound("Kategorija nije pronađena."))
		return
	}
	if errors.Is(e, repository.ErrCategoryInUse) {
		response.WriteError(w, &response.APIError{Status: 409, Code: "category_in_use", Message: "Kategorija se koristi i ne može biti obrisana."})
		return
	}
	if e != nil {
		response.WriteError(w, e)
		return
	}
	response.Write(w, 200, map[string]bool{"deleted": true})
}
func (h *Handler) Inquiries(w http.ResponseWriter, r *http.Request) {
	p, e := h.store.ListInquiries(r.Context(), models.InquiryFilter{Status: r.URL.Query().Get("status"), Page: queryInt(r, "page", 1), PageSize: queryInt(r, "pageSize", 20)})
	if e != nil {
		response.WriteError(w, e)
		return
	}
	response.Write(w, 200, p)
}
func (h *Handler) Inquiry(w http.ResponseWriter, r *http.Request) {
	i, e := h.store.InquiryByID(r.Context(), r.PathValue("id"))
	if errors.Is(e, repository.ErrNotFound) {
		response.WriteError(w, response.NotFound("Upit nije pronađen."))
		return
	}
	if e != nil {
		response.WriteError(w, e)
		return
	}
	response.Write(w, 200, i)
}
func (h *Handler) InquiryStatus(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Status string `json:"status"`
	}
	if e := decode(w, r, &b); e != nil {
		response.WriteError(w, e)
		return
	}
	if b.Status != "new" && b.Status != "read" && b.Status != "contacted" && b.Status != "closed" {
		response.WriteError(w, response.BadRequest(map[string]string{"status": "Nepoznat status."}))
		return
	}
	i, e := h.store.UpdateInquiryStatus(r.Context(), r.PathValue("id"), b.Status)
	if errors.Is(e, repository.ErrNotFound) {
		response.WriteError(w, response.NotFound("Upit nije pronađen."))
		return
	}
	if e != nil {
		response.WriteError(w, e)
		return
	}
	response.Write(w, 200, i)
}
