package services

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lunjae/jasen-jela-web/backend/internal/models"
)

var ErrInvalidImage = errors.New("invalid image")

// CloudinaryUploader keeps Cloudinary credentials on the server and performs
// signed uploads. Firestore only receives the returned URL and public_id.
type CloudinaryUploader struct {
	CloudName string
	APIKey    string
	APISecret string
	Client    *http.Client
	BaseURL   string
	Now       func() time.Time
}

type cloudinaryResponse struct {
	SecureURL string `json:"secure_url"`
	PublicID  string `json:"public_id"`
	Error     *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (u CloudinaryUploader) Upload(r *http.Request, productID string) ([]models.ProductImage, error) {
	r.Body = http.MaxBytesReader(nil, r.Body, 81<<20)
	if err := r.ParseMultipartForm(81 << 20); err != nil {
		return nil, fmt.Errorf("%w: najviše 8 slika od po 10 MB", ErrInvalidImage)
	}
	files := append([]*multipart.FileHeader{}, r.MultipartForm.File["images"]...)
	files = append(files, r.MultipartForm.File["image"]...)
	if len(files) < 1 || len(files) > 8 {
		return nil, fmt.Errorf("%w: izaberite od 1 do 8 slika", ErrInvalidImage)
	}
	alt := strings.TrimSpace(r.FormValue("alt"))
	if len(alt) > 200 {
		return nil, fmt.Errorf("%w: alternativni tekst je predugačak", ErrInvalidImage)
	}
	isPrimary := false
	var err error
	if raw := r.FormValue("isPrimary"); raw != "" {
		isPrimary, err = strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("%w: isPrimary nije ispravan", ErrInvalidImage)
		}
	}
	uploaded := make([]models.ProductImage, 0, len(files))
	for _, header := range files {
		image, err := u.uploadFile(r, productID, header, alt, isPrimary && len(uploaded) == 0)
		if err != nil {
			for _, completed := range uploaded {
				_ = u.Delete(r, completed.PublicID)
			}
			return nil, err
		}
		uploaded = append(uploaded, image)
	}
	return uploaded, nil
}

func (u CloudinaryUploader) uploadFile(r *http.Request, productID string, header *multipart.FileHeader, alt string, isPrimary bool) (models.ProductImage, error) {
	if header.Size < 1 {
		return models.ProductImage{}, fmt.Errorf("%w: slika je prazna", ErrInvalidImage)
	}
	if header.Size > 10<<20 {
		return models.ProductImage{}, fmt.Errorf("%w: slika može imati najviše 10 MB", ErrInvalidImage)
	}
	file, err := header.Open()
	if err != nil {
		return models.ProductImage{}, fmt.Errorf("%w: slika nije čitljiva", ErrInvalidImage)
	}
	defer file.Close()
	headerBytes := make([]byte, 512)
	n, readErr := file.Read(headerBytes)
	if readErr != nil && readErr != io.EOF {
		return models.ProductImage{}, fmt.Errorf("%w: slika nije čitljiva", ErrInvalidImage)
	}
	if n == 0 {
		return models.ProductImage{}, fmt.Errorf("%w: slika je prazna", ErrInvalidImage)
	}
	_, _ = file.Seek(0, io.SeekStart)
	contentType := http.DetectContentType(headerBytes[:n])
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		return models.ProductImage{}, fmt.Errorf("%w: dozvoljeni formati su JPG, PNG i WebP", ErrInvalidImage)
	}

	publicID := fmt.Sprintf("jasen-jela/products/%s/%s", productID, uuid.NewString())
	extension := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}[contentType]
	result, err := u.uploadOne(r, file, "upload"+extension, contentType, publicID)
	if err != nil {
		return models.ProductImage{}, err
	}
	return models.ProductImage{URL: result.SecureURL, PublicID: result.PublicID, Alt: alt, IsPrimary: isPrimary}, nil
}

func (u CloudinaryUploader) uploadOne(r *http.Request, src io.Reader, filename, contentType, publicID string) (cloudinaryResponse, error) {
	timestamp := strconv.FormatInt(u.now().Unix(), 10)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return cloudinaryResponse{}, fmt.Errorf("čuvanje slike nije uspelo")
	}
	if _, err = io.Copy(part, src); err != nil {
		return cloudinaryResponse{}, fmt.Errorf("čuvanje slike nije uspelo")
	}
	_ = writer.WriteField("api_key", u.APIKey)
	_ = writer.WriteField("timestamp", timestamp)
	_ = writer.WriteField("public_id", publicID)
	_ = writer.WriteField("signature", u.signature("public_id="+publicID+"&timestamp="+timestamp))
	if err = writer.Close(); err != nil {
		return cloudinaryResponse{}, fmt.Errorf("čuvanje slike nije uspelo")
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, u.endpoint("upload"), &body)
	if err != nil {
		return cloudinaryResponse{}, fmt.Errorf("čuvanje slike nije uspelo")
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := u.client().Do(req)
	if err != nil {
		return cloudinaryResponse{}, fmt.Errorf("Cloudinary nije dostupan")
	}
	defer resp.Body.Close()
	var result cloudinaryResponse
	if err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
		return result, fmt.Errorf("neispravan odgovor servisa za slike")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || result.Error != nil || result.SecureURL == "" || result.PublicID == "" {
		return result, fmt.Errorf("Cloudinary upload nije uspeo")
	}
	_ = contentType // Cloudinary determines and validates the uploaded resource type.
	return result, nil
}

func (u CloudinaryUploader) Delete(r *http.Request, publicID string) error {
	timestamp := strconv.FormatInt(u.now().Unix(), 10)
	values := url.Values{
		"api_key":   {u.APIKey},
		"timestamp": {timestamp},
		"public_id": {publicID},
		"signature": {u.signature("public_id=" + publicID + "&timestamp=" + timestamp)},
	}
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, u.endpoint("destroy"), strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("brisanje slike nije uspelo")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := u.client().Do(req)
	if err != nil {
		return fmt.Errorf("Cloudinary nije dostupan")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("brisanje slike nije uspelo")
	}
	var result struct {
		Result string `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil || result.Error != nil || (result.Result != "ok" && result.Result != "not found") {
		return fmt.Errorf("brisanje slike nije uspelo")
	}
	return nil
}

func (u CloudinaryUploader) signature(params string) string {
	sum := sha1.Sum([]byte(params + u.APISecret))
	return hex.EncodeToString(sum[:])
}

func (u CloudinaryUploader) endpoint(action string) string {
	base := strings.TrimRight(u.BaseURL, "/")
	if base == "" {
		base = "https://api.cloudinary.com"
	}
	return fmt.Sprintf("%s/v1_1/%s/image/%s", base, url.PathEscape(u.CloudName), action)
}

func (u CloudinaryUploader) client() *http.Client {
	if u.Client != nil {
		return u.Client
	}
	return &http.Client{Timeout: 30 * time.Second}
}

func (u CloudinaryUploader) now() time.Time {
	if u.Now != nil {
		return u.Now()
	}
	return time.Now()
}
