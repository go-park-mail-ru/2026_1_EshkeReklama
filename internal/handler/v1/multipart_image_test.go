package v1

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"testing"
)

func multipartAvatarHeader(t *testing.T, filename, contentType string, body []byte) *multipart.FileHeader {
	t.Helper()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("avatar", filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err = part.Write(body); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err = w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	req := httptest.NewRequest("POST", "/", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if err = req.ParseMultipartForm(maxAvatarSize); err != nil {
		t.Fatalf("ParseMultipartForm: %v", err)
	}

	fh := req.MultipartForm.File["avatar"][0]
	fh.Header.Set("Content-Type", contentType)
	return fh
}

func TestParseAndValidateImage_SuccessAndDefaults(t *testing.T) {
	file := multipartAvatarHeader(t, "avatar", "image/png", []byte("pngdata"))

	uploaded, err := ParseAndValidateImage(file)
	if err != nil {
		t.Fatalf("ParseAndValidateImage: %v", err)
	}
	if string(uploaded.Data) != "pngdata" {
		t.Fatalf("unexpected data: %q", string(uploaded.Data))
	}
	if uploaded.Ext != ".png" {
		t.Fatalf("expected .png got %s", uploaded.Ext)
	}
}

func TestParseAndValidateImage_UnsupportedType(t *testing.T) {
	file := multipartAvatarHeader(t, "avatar.gif", "image/gif", []byte("gifdata"))

	if _, err := ParseAndValidateImage(file); err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseAndValidateImage_TooLargeAndNil(t *testing.T) {
	if _, err := ParseAndValidateImage(nil); err == nil {
		t.Fatalf("expected nil-file error")
	}

	file := multipartAvatarHeader(t, "avatar.jpg", "image/jpeg", []byte("small"))
	file.Size = maxAvatarSize + 1
	if _, err := ParseAndValidateImage(file); err == nil {
		t.Fatalf("expected size error")
	}
}
