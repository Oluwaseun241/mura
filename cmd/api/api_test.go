package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func createMultipartForm(imageData []byte) (*bytes.Buffer, string) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("image", "test.jpg")
	part.Write(imageData)
	writer.Close()
	return body, writer.FormDataContentType()
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
	imageData := []byte{}
	body, contentType := createMultipartForm(imageData)

	req := httptest.NewRequest(http.MethodPost, "/detect", body)
	req.Header.Set(echo.HeaderContentType, contentType)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, IngredientHandler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)
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
