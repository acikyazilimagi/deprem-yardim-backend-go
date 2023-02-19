package search

import (
	"context"
	"os"

	log "github.com/acikkaynak/backend-api-go/pkg/logger"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type index[T any] struct {
	connStr string
	name    string
}

func NewIndex[T any](name string) *index[T] {
	connStr := os.Getenv("ELASTIC_CONN_STR")

	if connStr == "" {
		log.Logger().Panic("ELASTIC_CONN_STR env variable must be set")
	}

	return &index[T]{
		connStr: connStr,
		name:    name,
	}
}

func (i *index[T]) Bulk(ctx context.Context, items []Item[T]) error {
	var payload []byte

	for _, item := range items {
		payload = append(payload, []byte(`{"index":{"_index" : "`+i.name+`", "_id":"`+item.Id+`"}}`)...)
		payload = append(payload, '\n')
		source, _ := jsoniter.Marshal(item.Source)
		payload = append(payload, source...)
		payload = append(payload, '\n')
	}

	req := fasthttp.AcquireRequest()
	req.SetBody(payload)
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")
	req.SetRequestURI(i.connStr + "/_bulk")
	res := fasthttp.AcquireResponse()

	deadline, _ := ctx.Deadline()

	if err := fasthttp.DoDeadline(req, res, deadline); err != nil {
		return err
	}

	fasthttp.ReleaseRequest(req)

	body := res.Body()
	var response map[string]interface{}

	if err := jsoniter.Unmarshal(body, &response); err != nil {
		return err
	}

	log.Logger().Info("elastic bulk write response", zap.Any("response", response))

	fasthttp.ReleaseResponse(res)

	return nil
}

func (i *index[T]) Search(ctx context.Context, query map[string]interface{}) (*Result[T], error) {
	payload, _ := jsoniter.Marshal(query)

	req := fasthttp.AcquireRequest()
	req.SetBody(payload)
	req.Header.SetMethod("GET")
	req.Header.SetContentType("application/json")
	req.SetRequestURI(i.connStr + "/" + i.name + "/_search")
	res := fasthttp.AcquireResponse()

	deadline, _ := ctx.Deadline()

	if err := fasthttp.DoDeadline(req, res, deadline); err != nil {
		return nil, err
	}

	fasthttp.ReleaseRequest(req)

	body := res.Body()
	var response Result[T]

	if err := jsoniter.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	fasthttp.ReleaseResponse(res)

	return &response, nil
}
