package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}
func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	city := "moscow"
	totalCafes := len(cafeList[city])
	requests := []struct {
		count int
		want  int
	}{
		{1, 1},
		{2, 2},
		{100, totalCafes},
	}

	for _, test := range requests {

		url := fmt.Sprintf("/cafe?city=%s&count=%d", city, test.count)
		response := httptest.NewRecorder()
		request := httptest.NewRequest("GET", url, nil)

		handler.ServeHTTP(response, request)
		require.Equal(t, http.StatusOK, response.Code)

		body := strings.TrimSpace(response.Body.String())
		cafes := strings.Split(body, ",")

		assert.Equal(t, test.want, len(cafes),
			"Для count=%d ожидалось %d кафе", test.count, test.want)
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	requests := []struct {
		search    string // передаваемое значение search
		wantCount int    // ожидаемое количество кафе в ответе
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}
	for _, test := range requests {
		url := fmt.Sprintf("/cafe?city=moscow&search=%s", test.search)

		response := httptest.NewRecorder()
		requests := httptest.NewRequest("GET", url, nil)

		handler.ServeHTTP(response, requests)
		require.Equal(t, http.StatusOK, response.Code)

		body := strings.TrimSpace(response.Body.String())
		if test.wantCount == 0 {
			assert.Empty(t, body,
				"Для поиска='%s' должен быть пустой ответ", test.search)
			continue
		}
		cafe := strings.Split(body, ",")
		assert.Equal(t, test.wantCount, len(cafe))
		searchLower := strings.ToLower(test.search)
		for _, cafe := range cafe {
			cafeLower := strings.ToLower(cafe)
			assert.True(t, strings.Contains(cafeLower, searchLower),
				"Кафе '%s' должно содержать '%s'", cafe, test.search)
		}
	}

}
