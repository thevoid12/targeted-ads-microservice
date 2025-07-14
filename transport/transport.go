package transport

import (
	"context"
	"encoding/json"
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
	return req, err
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}
