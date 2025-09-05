package models

import (
	"github.com/r3dpixel/toolkit/gjsonx"
	"github.com/tidwall/gjson"
)

var EmptyJsonResponse = JsonResponse{
	Metadata:         gjsonx.Empty,
	BookResponses:    nil,
	AuxBookResponses: nil,
}

type JsonResponse struct {
	Metadata         gjson.Result
	BookResponses    []gjson.Result
	AuxBookResponses []gjson.Result
}

func (gJsonData *JsonResponse) BookCount() int {
	return len(gJsonData.BookResponses) + len(gJsonData.AuxBookResponses)
}
