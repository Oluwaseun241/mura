package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func createMultipartForm(imageData []byte) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("images", "data.jpeg")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(imageData); err != nil {
		return nil, "", fmt.Errorf("failed to write image data: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	return body, writer.FormDataContentType(), nil
}

func mockClassifyImage(fileBytes []byte) (string, error) {
	return "ingredient", nil
}

// func TestFoodHandler(t *testing.T) {
// 	e := echo.New()
// 	imageData := []byte{}
// 	body, contentType := createMultipartForm(imageData)
//
// 	req := httptest.NewRequest(http.MethodPost, "/detect-food", body)
// 	req.Header.Set(echo.HeaderContentType, contentType)
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)
//
// 	internal.ClassifyImage = mockClassifyImage()
// }

func TestIngredientHandler(t *testing.T) {
	e := echo.New()

	filePath := "../../tmp/raw.jpeg"
	src, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open image file: %v", err)
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		t.Fatalf("Failed to read image file: %v", err)
	}

	body, contentType, err := createMultipartForm(fileBytes)
	if err != nil {
		t.Fatalf("Failed to create multipart form: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/detect", body)
	req.Header.Set(echo.HeaderContentType, contentType)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, IngredientHandler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		assert.True(t, response["status"].(bool))
		assert.NotEmpty(t, response["data"])
	}
}

func TestRecipeHandler(t *testing.T) {
	e := echo.New()
	ingredients := []string{"tomato", "cheese", "basil"}
	data := map[string]interface{}{
		"ingredients": ingredients,
		"dish":        "pizza",
	}
	body, _ := json.Marshal(data)

	req := httptest.NewRequest(http.MethodPost, "/recipe", strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, RecipeHandler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)
		assert.True(t, response["status"].(bool))
		assert.NotEmpty(t, response["data"])
	}
}
