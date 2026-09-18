package fincloudapi

import (
	"context"
	"net/http"
)

func (c *Client) InquiryLoan(
	ctx context.Context,
	payload LoanInquiryRequest,
) (*LoanInquiryResponse, error) {
	return doAPI[LoanInquiryResponse](
		c,
		ctx,
		http.MethodPost,
		"/inquiry/detail/loan",
		payload,
		nil,
	)
}

func (c *Client) DisburseLoan(
	ctx context.Context,
	payload LoanDisbursementRequest,
) (*LoanDisbursementResponse, error) {
	return doAPI[LoanDisbursementResponse](
		c,
		ctx,
		http.MethodPost,
		"/disbursement",
		payload,
		nil,
	)
}

func (c *Client) TerminateLoan(
	ctx context.Context,
	payload LoanTerminationRequest,
) (*LoanTerminationResponse, error) {
	return doAPI[LoanTerminationResponse](
		c,
		ctx,
		http.MethodPost,
		"/loan/earlytermination/",
		payload,
		nil,
	)
}

func (c *Client) RepayLoan(
	ctx context.Context,
	payload LoanRepaymentRequest,
) (*LoanRepaymentResponse, error) {
	return doAPI[LoanRepaymentResponse](
		c,
		ctx,
		http.MethodPost,
		"/loan/repayment/",
		payload,
		nil,
	)
}

func (c *Client) InquiryEarlyTermination(
	ctx context.Context,
	payload InquiryETOrRepaymentRequest,
) (*InquiryEarlyTerminationResponse, error) {
	return doAPI[InquiryEarlyTerminationResponse](
		c,
		ctx,
		http.MethodPost,
		"/loan/earlytermination/status",
		payload,
		nil,
	)
}

func (c *Client) InquiryRepayment(
	ctx context.Context,
	payload InquiryETOrRepaymentRequest,
) (*InquiryLoanRepaymentResponse, error) {
	return doAPI[InquiryLoanRepaymentResponse](
		c,
		ctx,
		http.MethodPost,
		"/loan/repayment/status",
		payload,
		nil,
	)
}

func (c *Client) InquiryPortfolioLoanByCIFNo(
	ctx context.Context,
	cifNo string,
) ([]PortfolioLoan, error) {
	return doAPIList[PortfolioLoan](
		c,
		ctx,
		http.MethodPost,
		"/accounts/portfolioLoanByCIForNationalID",
		nil,
		func(req *http.Request) error {
			q := req.URL.Query()
			addNonEmptyQuery(q, "cifNo", cifNo)
			req.URL.RawQuery = q.Encode()
			return nil
		},
	)
}

func (c *Client) InquiryLoanProducts(
	ctx context.Context,
) ([]LoanProduct, error) {
	return doAPIList[LoanProduct](
		c,
		ctx,
		http.MethodPost,
		"/loan/product",
		nil,
		nil,
	)
}
