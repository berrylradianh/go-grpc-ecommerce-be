package utils

import "github.com/berrylradianh/go-grpc-ecommerce-be/pb/common"

func SuccessResponse(message string) *common.BaseResponse {
	return &common.BaseResponse{
		StatusCode: 200,
		Message:    message,
		IsError:    false,
	}
}

func ValidationErrorResponse(validationErrors []*common.ValidationError) *common.BaseResponse {
	return &common.BaseResponse{
		StatusCode:       400,
		Message:          "Validation Error",
		IsError:          true,
		ValidationErrors: validationErrors,
	}
}

func BadRequestResponse(message string) *common.BaseResponse {
	return &common.BaseResponse{
		StatusCode:       400,
		Message:          message,
		IsError:          true,
		ValidationErrors: nil,
	}
}
