package itwofactorservice

import (
	"context"
	"net/http"
)

// QueryTimeResponse describes the response for the QueryTime method.
type QueryTimeResponse struct {
	ServerTime int64 `json:"server_time,string"`
}

// QueryTime queries the server time.
func (a *ITwoFactorService) QueryTime(ctx context.Context) (*QueryTimeResponse, error) {
	var resBody struct {
		Response QueryTimeResponse `json:"response"`
	}
	if err := a.transport.Call(ctx, http.MethodPost, "ITwoFactorService", "QueryTime", 1, nil, &resBody); err != nil {
		return nil, err
	}
	return &resBody.Response, nil
}
