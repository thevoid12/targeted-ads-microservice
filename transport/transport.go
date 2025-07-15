package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"targetad/endpoint"
	"targetad/pkg/target/model"

	httptransport "github.com/go-kit/kit/transport/http"
)

func NewHTTPHandler() http.Handler {
	m := http.NewServeMux()

	m.Handle("/v1/delivery", httptransport.NewServer(
		endpoint.MakeDeliveryServiceEndpoint(),
		decodeDeliveryAdsRequest,
		encodeResponse,
	))

	return m
}

func decodeDeliveryAdsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req model.DeliveryServiceRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, fmt.Errorf("error decoding the input request: %s", err)
	}
	return req, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}
