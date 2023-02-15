package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/acikkaynak/backend-api-go/feeds"
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
	jsonBytes, err := json.Marshal(NeedsRequest{
		Inputs: []string{fullText},
	})

	req, err := http.NewRequest("POST", os.Getenv("NEEDS_RESOLVER_API_URL"), bytes.NewReader(jsonBytes))
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not prepare http request NeedsMessagePayload error message %s error %s", fullText, err.Error())
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+os.Getenv("NEEDS_RESOLVER_API_KEY"))
	req.Header.Add("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "could not get response NeedsMessagePayload feedID %d status %d", feedID, resp.StatusCode)
		return nil, err
	}

	needsResp := &NeedsResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&needsResp); err != nil {
		fmt.Fprintf(os.Stderr, "could not get decode response NeedsMessagePayload feedID %d err %s", feedID, err.Error())
		return nil, err
	}

	needs := make([]feeds.NeedItem, 0)
	if len(needsResp.Response) == 0 {
		fmt.Fprintf(os.Stderr, "no data found on response NeedsMessagePayload feedID %d", feedID)
		// ret empty
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
