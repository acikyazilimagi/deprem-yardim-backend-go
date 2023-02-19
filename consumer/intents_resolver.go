package consumer

import (
	"bytes"
	"context"
	"fmt"

	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/acikkaynak/backend-api-go/pkg/logger"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

type IntentRequest struct {
	Inputs string `json:"inputs"`
}

type IntentResponse struct {
	Results []Intent
}

type Intent []struct {
	Label string  `json:"label"`
	Score float64 `json:"score"`
}

type IntentMessagePayload struct {
	FeedID          int64  `json:"id"`
	FullText        string `json:"full_text"`
	ResolvedAddress string `json:"resolved_address"`
}

func (consumer *Consumer) intentResolveHandle(message *sarama.ConsumerMessage, session sarama.ConsumerGroupSession) {
	var messagePayload IntentMessagePayload
	if err := jsoniter.Unmarshal(message.Value, &messagePayload); err != nil {
		log.Logger().Error("deserialization IntentMessagePayload error", zap.String("message", string(message.Value)), zap.Error(err))
		session.MarkMessage(message, "")
		session.Commit()
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	intents, err := sendIntentResolveRequest(messagePayload.FullText, messagePayload.FeedID)
	if err != nil {
		if err.Error() == "alakasiz veri" {
			if err := consumer.repo.DeleteFeedLocation(ctx, messagePayload.FeedID); err != nil {
				log.Logger().Error("", zap.Error(err))
			}
		}
		log.Logger().Error("", zap.Error(err))
		session.MarkMessage(message, "")
		session.Commit()
		return
	}

	needs, err := sendNeedsResolveRequest(messagePayload.FullText, messagePayload.FeedID)
	if err != nil {
		session.MarkMessage(message, "")
		session.Commit()
		return
	}

	needsForDuplication := make([]string, 0)
	for _, n := range needs {
		needsForDuplication = append(needsForDuplication, n.Label)
	}
	isDuplicate, err := checkDuplication(DuplicationRequest{
		Address: messagePayload.ResolvedAddress,
		Intents: strings.Split(intents, ","),
		Needs:   needsForDuplication,
	})

	if err != nil {
		log.Logger().Error("could not get duplicate response", zap.Error(err))
		isDuplicate = false
	}

	if isDuplicate {
		if err := consumer.repo.DeleteFeedLocation(ctx, messagePayload.FeedID); err != nil {
			log.Logger().Error("could not delete feed location after duplication request",
				zap.Int64("entry_id", messagePayload.FeedID),
				zap.Error(err))
		}
		return
	}

	if err := consumer.repo.UpdateLocationIntentAndNeeds(ctx, messagePayload.FeedID, intents, needs); err != nil {
		log.Logger().Error("error updating feed entry, location intent and needs",
			zap.Error(err), zap.String("payload", string(message.Value)))
		return
	}

	session.MarkMessage(message, "")
	session.Commit()
}

func sendIntentResolveRequest(fullText string, feedID int64) (string, error) {
	jsonBytes, err := jsoniter.Marshal(IntentRequest{
		Inputs: fullText,
	})

	req, err := http.NewRequest("POST", os.Getenv("INTENT_RESOLVER_API_URL"), bytes.NewReader(jsonBytes))
	if err != nil {
		log.Logger().Error("could not prepare http request IntentMessagePayload", zap.String("fullText", fullText), zap.Error(err))
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+os.Getenv("INTENT_RESOLVER_API_KEY"))
	req.Header.Add("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		log.Logger().Error("could not get response IntentMessagePayload", zap.Int64("feedID", feedID), zap.Int("statusCode", resp.StatusCode))
		return "", err
	}

	intentResp := &IntentResponse{}
	if err := jsoniter.NewDecoder(resp.Body).Decode(&intentResp.Results); err != nil {
		log.Logger().Error("could not get decode response IntentMessagePayload", zap.Int64("feedID", feedID), zap.Error(err))
		return "", err
	}

	if len(intentResp.Results) == 0 {
		log.Logger().Error("no data found on response IntentMessagePayload", zap.Int64("feedID", feedID))
		return "", nil
	}

	intents := make([]string, 0)
	for _, val := range intentResp.Results[0] {
		if val.Score >= 0.4 {
			if val.Label == "Alakasiz" && val.Score >= 0.7 {
				return "", fmt.Errorf("alakasiz veri")
			}
			intents = append(intents, strings.ToLower(val.Label))
		}
	}

	return strings.Join(intents, ","), nil
}

func checkDuplication(payload DuplicationRequest) (bool, error) {
	jsonBytes, err := jsoniter.Marshal(payload)

	req, err := http.NewRequest("POST", os.Getenv("DUPLICATION_API_URL"), bytes.NewReader(jsonBytes))
	if err != nil {
		log.Logger().Error("could not prepare http request DuplicationRequest", zap.Error(err))
		return false, err
	}
	req.Header.Add("Authorization", "Bearer "+os.Getenv("DUPLICATION_API_KEY"))
	req.Header.Add("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		log.Logger().Error("could not get response DuplicationRequest", zap.Error(err))
		return false, err
	}

	duplicationResp := &DuplicationResponse{}
	if err := jsoniter.NewDecoder(resp.Body).Decode(&duplicationResp); err != nil {
		log.Logger().Error("could not get decode response DuplicationRequest", zap.Error(err))
		return false, err
	}

	return duplicationResp.IsDuplicate, nil
}
