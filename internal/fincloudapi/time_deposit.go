package fincloudapi

import (
	"context"
	"net/http"
)

func (c *Client) InquiryTimeDepositByCIF(
	ctx context.Context,
	cifNo string,
) ([]TimeDepositDetail, error) {
	return doAPIList[TimeDepositDetail](
		c,
		ctx,
		http.MethodGet,
		"/account/deposit/list",
		nil,
		func(req *http.Request) error {
			q := req.URL.Query()
			addNonEmptyQuery(q, "cifNo", cifNo)
			req.URL.RawQuery = q.Encode()

			return nil
		},
	)
}

func (c *Client) InquiryTimeDepositByCIF2(
	ctx context.Context,
	cifNo string,
) ([]TimeDepositDetail2, error) {
	return doAPIList[TimeDepositDetail2](
		c,
		ctx,
		http.MethodGet,
		"/v2/account/deposit/list",
		nil,
		func(req *http.Request) error {
			q := req.URL.Query()
			addNonEmptyQuery(q, "cifNo", cifNo)
			req.URL.RawQuery = q.Encode()

			return nil
		},
	)
}

func (c *Client) InquiryTimeDepositProducts(
	ctx context.Context,
) ([]TimeDepositProduct, error) {
	return doAPIList[TimeDepositProduct](
		c,
		ctx,
		http.MethodGet,
		"/deposit/product/list",
		nil,
		nil,
	)
}
