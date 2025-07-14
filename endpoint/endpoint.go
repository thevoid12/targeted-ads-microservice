package endpoint

import (
	"context"
	"targetad/pkg/target"
	"targetad/pkg/target/model"

	"github.com/go-kit/kit/endpoint"
)

func MakeDeliveryServiceEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(model.DeliveryServiceRequest)
		v, err := target.DeliveryService(ctx, &req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}
