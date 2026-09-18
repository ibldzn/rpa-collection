package fincloudapi

import (
	"context"
	"net/http"
)

func (c *Client) TransferGlToGl(ctx context.Context, payload GlToGlRequest) (*GlToGlResponse, error) {
	return doAPI[GlToGlResponse](
		c,
		ctx,
		http.MethodPost,
		"/trx/transfer/gl-to-gl",
		payload,
		nil,
	)
}

func (c *Client) TransferGlToSaving(ctx context.Context, payload GlToSavingRequest) (*GlToSavingResponse, error) {
	return doAPI[GlToSavingResponse](
		c,
		ctx,
		http.MethodPost,
		"/trx/transfer/gl",
		payload,
		nil,
	)
}

func (c *Client) TransferOverbooking(ctx context.Context, payload TransferOverbookingRequest) (*baseResponse, error) {
	payload.TrxType = "TL01"
	payload.TermType = "6014"
	payload.TermID = "FINCLOUD"

	return doAPI[baseResponse](
		c,
		ctx,
		http.MethodPost,
		"/trx/transfer",
		payload,
		nil,
	)
}
