package fincloudapi

import (
	"context"
	"fmt"
	"net/http"
)

func (c *Client) InquirySavingBalance(
	ctx context.Context,
	accountNumber string,
) (*SavingBalanceInquiryResponse, error) {
	return doAPI[SavingBalanceInquiryResponse](
		c,
		ctx,
		http.MethodGet,
		"/saving/inq/balance",
		nil,
		func(req *http.Request) error {
			q := req.URL.Query()
			addNonEmptyQuery(q, "accountNumber", accountNumber)
			req.URL.RawQuery = q.Encode()

			return nil
		},
	)
}

func (c *Client) InquirySavingBalance2(
	ctx context.Context,
	accountNumber string,
) (*SavingBalanceInquiry2Response, error) {
	return doAPI[SavingBalanceInquiry2Response](
		c,
		ctx,
		http.MethodGet,
		"/account/balance",
		nil,
		func(req *http.Request) error {
			q := req.URL.Query()
			addNonEmptyQuery(q, "account", accountNumber)
			req.URL.RawQuery = q.Encode()

			return nil
		},
	)
}

func (c *Client) InquirySavingTransactionHistory(
	ctx context.Context,
	payload SavingStatementInquiryRequest,
) (*InquiryAccountStatementsResponse, error) {
	return doAPI[InquiryAccountStatementsResponse](
		c,
		ctx,
		http.MethodGet,
		"/account/statement",
		nil,
		func(req *http.Request) error {
			q := req.URL.Query()
			addNonEmptyQuery(q, "accountNo", payload.AccountNumber)
			addNonEmptyQuery(q, "startDate", payload.StartDate)
			addNonEmptyQuery(q, "endDate", payload.EndDate)
			addNonEmptyQuery(q, "rowPerPage", fmt.Sprintf("%d", payload.RowPerPage))
			addNonEmptyQuery(q, "index", fmt.Sprintf("%d", payload.Index))
			req.URL.RawQuery = q.Encode()

			return nil
		},
	)
}

func (c *Client) InquirySavingBalanceByAccOrAltNumber(
	ctx context.Context,
	accountNumber string,
	isAltNumber bool,
) (*SavingBalanceInquiryResponse, error) {
	key := "accountNumber"
	if isAltNumber {
		key = "altNumber"
	}

	body := map[string]string{
		key: accountNumber,
	}

	return doAPI[SavingBalanceInquiryResponse](
		c,
		ctx,
		http.MethodPost,
		"/saving/account/balanceByAccNoOrAltNo",
		body,
		nil,
	)
}

func (c *Client) InquirySavingListByCIF(
	ctx context.Context,
	cifNo string,
) ([]SavingDetail, error) {
	return doAPIList[SavingDetail](
		c,
		ctx,
		http.MethodGet,
		"/account/saving/list",
		nil,
		func(req *http.Request) error {
			q := req.URL.Query()
			addNonEmptyQuery(q, "cifNo", cifNo)
			req.URL.RawQuery = q.Encode()
			return nil
		},
	)
}

func (c *Client) InquirySavingProducts(
	ctx context.Context,
) ([]SavingProduct, error) {
	return doAPIList[SavingProduct](
		c,
		ctx,
		http.MethodGet,
		"/saving/product/list",
		nil,
		nil,
	)
}

func (c *Client) CreateSaving(
	ctx context.Context,
	payload CreateSavingRequest,
) (*CreateSavingResponse, error) {
	return doAPI[CreateSavingResponse](
		c,
		ctx,
		http.MethodPost,
		"/account/create/saving",
		payload,
		nil,
	)
}

func (c *Client) BlockSaving(
	ctx context.Context,
	payload BlockSavingRequest,
) (*BlockSavingResponse, error) {
	return doAPI[BlockSavingResponse](
		c,
		ctx,
		http.MethodPost,
		"/account/blocking",
		payload,
		nil,
	)
}

func (c *Client) UnblockSaving(
	ctx context.Context,
	payload UnblockSavingRequest,
) (*UnblockSavingResponse, error) {
	return doAPI[UnblockSavingResponse](
		c,
		ctx,
		http.MethodPost,
		"/account/unBlocking",
		payload,
		nil,
	)
}
