package consumer

import (
	"bytes"
	"net/http"
	"os"
	"strings"

	"github.com/acikkaynak/backend-api-go/feeds"
	log "github.com/acikkaynak/backend-api-go/pkg/logger"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

type NeedsRequest struct {
	Inputs []string `json:"inputs"`
}

type NeedsResponse struct {
	Response []struct {
		String    []string `json:"string"`
		Processed struct {
			Intent             []string `json:"intent"`
			DetailedIntentTags []string `json:"detailed_intent_tags"`
		} `json:"processed"`
	} `json:"response"`
}

func sendNeedsResolveRequest(fullText string, feedID int64) ([]feeds.NeedItem, error) {
	jsonBytes, err := jsoniter.Marshal(NeedsRequest{
		Inputs: []string{fullText},
	})

	req, err := http.NewRequest("POST", os.Getenv("NEEDS_RESOLVER_API_URL"), bytes.NewReader(jsonBytes))
	if err != nil {
		log.Logger().Error("could not prepare http request NeedsMessagePayload", zap.String("message", fullText), zap.Error(err))
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+os.Getenv("NEEDS_RESOLVER_API_KEY"))
	req.Header.Add("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		log.Logger().Error("could not get response NeedsMessagePayload", zap.Int64("feedID", feedID), zap.Int("statusCode", resp.StatusCode))
		return nil, err
	}

	needsResp := &NeedsResponse{}
	if err := jsoniter.NewDecoder(resp.Body).Decode(&needsResp); err != nil {
		log.Logger().Error("could not get decode response NeedsMessagePayload", zap.Int64("feedID", feedID), zap.Error(err))
		return nil, err
	}

	needs := make([]feeds.NeedItem, 0)
	if len(needsResp.Response) == 0 {
		log.Logger().Error("no data found on response NeedsMessagePayload", zap.Int64("feedID", feedID))
		return needs, nil
	}

	for _, tag := range needsResp.Response[0].Processed.DetailedIntentTags {
		needs = append(needs, feeds.NeedItem{
			Label:  strings.ToLower(tag),
			Status: true,
		})
	}

	return needs, nil
}
