package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestJsonResponse_BookCount(t *testing.T) {
	testCases := []struct {
		name              string
		jsonResponse      *JsonResponse
		expectedBookCount int
	}{
		{
			name:              "Both slices are nil",
			jsonResponse:      &JsonResponse{BookResponses: nil, AuxBookResponses: nil},
			expectedBookCount: 0,
		},
		{
			name:              "Both slices are empty",
			jsonResponse:      &JsonResponse{BookResponses: []gjson.Result{}, AuxBookResponses: []gjson.Result{}},
			expectedBookCount: 0,
		},
		{
			name: "Only BookResponses has items",
			jsonResponse: &JsonResponse{
				BookResponses:    []gjson.Result{gjson.Parse(`{}`), gjson.Parse(`{}`)},
				AuxBookResponses: nil,
			},
			expectedBookCount: 2,
		},
		{
			name: "Only AuxBookResponses has items",
			jsonResponse: &JsonResponse{
				BookResponses:    nil,
				AuxBookResponses: []gjson.Result{gjson.Parse(`{}`)},
			},
			expectedBookCount: 1,
		},
		{
			name: "Both slices have items",
			jsonResponse: &JsonResponse{
				BookResponses:    []gjson.Result{gjson.Parse(`{}`), gjson.Parse(`{}`)},
				AuxBookResponses: []gjson.Result{gjson.Parse(`{}`), gjson.Parse(`{}`), gjson.Parse(`{}`)},
			},
			expectedBookCount: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			count := tc.jsonResponse.BookCount()
			assert.Equal(t, tc.expectedBookCount, count)
		})
	}
}
