package api

import (
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
)

func printResponse(resp *genai.GenerateContentResponse) string {
	var result strings.Builder
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				result.WriteString(fmt.Sprintf("%v\n", part))
			}
		}
	}
	return result.String()
}

func removeDuplicates(elements []interface{}) []interface{} {
	encountered := map[string]bool{}
	result := []interface{}{}

	for _, v := range elements {
		strValue := fmt.Sprintf("%v", v)
		if !encountered[strValue] {
			encountered[strValue] = true
			result = append(result, v)
		}
	}

	return result
}
