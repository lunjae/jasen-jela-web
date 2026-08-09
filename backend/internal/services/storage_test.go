package services

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCloudinaryUpload(t *testing.T) {
	fixed := time.Unix(1_700_000_000, 0)
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1_1/demo/image/upload" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		publicID := r.FormValue("public_id")
		if !strings.HasPrefix(publicID, "jasen-jela/products/product-1/") {
			t.Fatalf("unexpected public id: %s", publicID)
		}
		u := CloudinaryUploader{APISecret: "secret"}
		expected := u.signature("public_id=" + publicID + "&timestamp=1700000000")
		if r.FormValue("signature") != expected || r.FormValue("api_key") != "key" {
			t.Fatal("signed Cloudinary fields are incorrect")
		}
		body := io.NopCloser(strings.NewReader(`{"secure_url":"https://res.cloudinary.com/demo/image/upload/example.webp","public_id":"` + publicID + `"}`))
		return &http.Response{StatusCode: http.StatusOK, Body: body, Header: make(http.Header)}, nil
	})}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, _ := mw.CreateFormFile("image", "../sample.png")
	// A valid PNG signature is enough for net/http content sniffing in this unit test.
	png, _ := hex.DecodeString("89504e470d0a1a0a0000000d49484452")
	_, _ = part.Write(png)
	_ = mw.WriteField("alt", "Model Elegance")
	_ = mw.WriteField("isPrimary", "true")
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/images", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	uploader := CloudinaryUploader{CloudName: "demo", APIKey: "key", APISecret: "secret", Client: client, Now: func() time.Time { return fixed }}
	images, err := uploader.Upload(req, "product-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 1 || images[0].URL == "" || !strings.HasPrefix(images[0].PublicID, "jasen-jela/products/product-1/") || !images[0].IsPrimary || images[0].Alt != "Model Elegance" {
		t.Fatalf("unexpected images: %#v", images)
	}
}

func TestCloudinaryUploadRejectsUnsupportedMIME(t *testing.T) {
	req := multipartUploadRequest(t, "image", "sample.txt", []byte("plain text"))
	_, err := (CloudinaryUploader{}).Upload(req, "product-1")
	if !errors.Is(err, ErrInvalidImage) {
		t.Fatalf("expected ErrInvalidImage, got %v", err)
	}
}

func TestCloudinaryUploadRejectsEmptyFile(t *testing.T) {
	req := multipartUploadRequest(t, "image", "empty.png", nil)
	_, err := (CloudinaryUploader{}).Upload(req, "product-1")
	if !errors.Is(err, ErrInvalidImage) {
		t.Fatalf("expected ErrInvalidImage, got %v", err)
	}
}

func TestCloudinaryUploadRejectsOversizedFile(t *testing.T) {
	content := make([]byte, (10<<20)+1)
	copy(content, []byte("\x89PNG\r\n\x1a\n"))
	req := multipartUploadRequest(t, "image", "large.png", content)
	_, err := (CloudinaryUploader{}).Upload(req, "product-1")
	if !errors.Is(err, ErrInvalidImage) {
		t.Fatalf("expected ErrInvalidImage, got %v", err)
	}
}

func multipartUploadRequest(t *testing.T, field, filename string, content []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile(field, filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err = mw.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/images", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func TestCloudinaryDelete(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1_1/demo/image/destroy" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = r.ParseForm()
		if r.FormValue("public_id") != "products/p/image" {
			t.Fatalf("unexpected public id: %s", r.FormValue("public_id"))
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"result":"ok"}`)), Header: make(http.Header)}, nil
	})}

	uploader := CloudinaryUploader{CloudName: "demo", APIKey: "key", APISecret: "secret", Client: client, Now: func() time.Time { return time.Unix(1_700_000_000, 0) }}
	if err := uploader.Delete(httptest.NewRequest(http.MethodDelete, "/", nil), "products/p/image"); err != nil {
		t.Fatal(err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
