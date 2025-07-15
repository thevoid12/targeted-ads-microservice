package endpoint

import (
	"context"
	"targetad/pkg/target"
	"targetad/pkg/target/model"

	"github.com/go-kit/kit/endpoint"
	validator "github.com/go-playground/validator/v10"
)

func MakeDeliveryServiceEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(model.DeliveryServiceRequest)
		validate := validator.New(validator.WithRequiredStructEnabled())

		// returns nil or ValidationErrors ( []FieldError )
		err := validate.Struct(req)
		if err != nil {
			return nil, err
		}
		v, err := target.DeliveryService(ctx, &req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}
