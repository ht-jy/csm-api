package handler

import (
	"context"
	"csm-api/entity"
	"encoding/json"
	"fmt"
	"net/http"
)

type ResultRole string

const (
	Success ResultRole = "Success"
	Failure ResultRole = "Failure"
)

type ErrDetailsRole string

const (
	ResponseEncodeError ErrDetailsRole = "Response Encode Error"
	InvalidToken        ErrDetailsRole = "Invalid Token or Not Found Token"
	BodyDataParseError  ErrDetailsRole = "Body Data Parse Error"
	InvalidUser         ErrDetailsRole = "Invalid User"
	TokenCreatedFail    ErrDetailsRole = "Token Created Fail"
	NotFoundParam       ErrDetailsRole = "Not Found Parameter"
	ParsingError        ErrDetailsRole = "Parsing Error"
	DataAddFailed       ErrDetailsRole = "Data Add Failed"
	DataModifyFailed    ErrDetailsRole = "Data Modify Failed"
	DataRemoveFailed    ErrDetailsRole = "Data Remove Failed"
	DataMergeFailed     ErrDetailsRole = "Data Merge Failed"
	CallApiFailed       ErrDetailsRole = "Call Api Failed"
)

type ErrResponse struct {
	Result         ResultRole     `json:"result"`
	Message        string         `json:"message"`
	Details        ErrDetailsRole `json:"details"`
	Error          string         `json:"error"`
	HttpStatusCode int            `json:"http-status-code"`
}

type Response struct {
	Result ResultRole `json:"result"`
	Values any        `json:"values"`
}

func RespondJSON(ctx context.Context, w http.ResponseWriter, body any, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	bodyBytes, err := json.Marshal(body)

	if err != nil {
		fmt.Printf("encode response error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		rsp := ErrResponse{
			Result:  Success,
			Message: http.StatusText(http.StatusInternalServerError),
			Details: ResponseEncodeError,
		}
		if err := json.NewEncoder(w).Encode(rsp); err != nil {
			fmt.Printf("write error response error: %v", err)
		}
		return
	}

	w.WriteHeader(status)
	if _, err = fmt.Fprintf(w, "%s", bodyBytes); err != nil {
		fmt.Printf("write response error: %v", err)
	}
}

func FailResponse(ctx context.Context, w http.ResponseWriter, err error) {
	//에러 로그 기록
	_ = entity.WriteErrorLog(ctx, err)

	RespondJSON(
		ctx,
		w,
		&ErrResponse{
			Result:         Failure,
			Message:        err.Error(),
			HttpStatusCode: http.StatusInternalServerError,
		},
		http.StatusOK)
}

func FailResponseMessage(ctx context.Context, w http.ResponseWriter, err error, message string) {
	//에러 로그 기록
	_ = entity.WriteErrorLog(ctx, err)

	RespondJSON(
		ctx,
		w,
		&ErrResponse{
			Result:         Failure,
			Error:          err.Error(),
			Message:        message,
			HttpStatusCode: http.StatusInternalServerError,
		},
		http.StatusOK)
}

func BadRequestResponse(ctx context.Context, w http.ResponseWriter) {
	RespondJSON(
		ctx,
		w,
		&ErrResponse{
			Result:         Failure,
			Message:        http.StatusText(http.StatusBadRequest),
			HttpStatusCode: http.StatusInternalServerError,
		},
		http.StatusOK)
}

func SuccessResponse(ctx context.Context, w http.ResponseWriter) {
	rsp := Response{
		Result: Success,
	}
	RespondJSON(ctx, w, rsp, http.StatusOK)
}

func SuccessValuesResponse(ctx context.Context, w http.ResponseWriter, values any) {
	rsp := Response{
		Result: Success,
		Values: values,
	}
	RespondJSON(ctx, w, rsp, http.StatusOK)
}
