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
	const city = "moscow"
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		count int
		want  int
	}{
		{count: 0, want: 0},
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: min(100, len(cafeList[city]))},
	}

	for _, testCase := range requests {
		target := fmt.Sprintf("/cafe?city=%s&count=%d", city, testCase.count)
		response := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, target, nil)
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code, "count=%d", testCase.count)

		body := strings.TrimSpace(response.Body.String())
		got := 0
		if body != "" {
			got = len(strings.Split(body, ","))
		}

		assert.Equal(t, testCase.want, got, "count=%d", testCase.count)
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		search    string
		wantCount int
	}{
		{search: "фасоль", wantCount: 0},
		{search: "кофе", wantCount: 2},
		{search: "вилка", wantCount: 1},
	}

	for _, testCase := range requests {
		target := fmt.Sprintf("/cafe?city=moscow&search=%s", testCase.search)
		response := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, target, nil)
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code, "search=%q", testCase.search)

		body := strings.TrimSpace(response.Body.String())
		var cafes []string
		if body != "" {
			cafes = strings.Split(body, ",")
		}

		assert.Equal(t, testCase.wantCount, len(cafes), "search=%q", testCase.search)

		search := strings.ToLower(testCase.search)
		for _, cafe := range cafes {
			assert.True(t, strings.Contains(strings.ToLower(cafe), search),
				"search=%q, кафе=%q", testCase.search, cafe)
		}
	}
}
