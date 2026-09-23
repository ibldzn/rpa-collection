package fincloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ibldzn/kolek-rpa/internal/fincloudapi"
)

const (
	autodebitRemovalInquiryPath = "/pinjaman/pendaftaranPenghapusanAutodebit/pembuatan/cari"
	autodebitRemovalSubmitPath  = "/pinjaman/pendaftaranPenghapusanAutodebit/pembuatan/pinjaman"
)

var operAccountPattern = regexp.MustCompile(`^[0-9]{3}000OPER$`)

type fincloudDate struct {
	Date string `json:"date"`
}

type autodebitRemovalLoan struct {
	Location                      string        `json:"lokasi"`
	AccountNumber                 string        `json:"id"`
	AgreementNumber               string        `json:"nopk"`
	CustomerName                  string        `json:"namanasabah"`
	AliasName                     string        `json:"aliasnama"`
	AccountStatus                 string        `json:"statusrekening"`
	DisbursementDate              *fincloudDate `json:"tgl_pencairan"`
	CreatedBy                     string        `json:"rec_dibuat_oleh"`
	AlternateNumber               string        `json:"noalt"`
	LoanType                      string        `json:"produk_jenispinjaman"`
	Product                       string        `json:"produkid"`
	ProductID                     string        `json:"idproduk"`
	Currency                      string        `json:"currency"`
	CreditLimit                   json.Number   `json:"plafondlimit"`
	LoanPrincipal                 json.Number   `json:"jmlpokok_pinjaman"`
	Term                          string        `json:"jangkawaktu"`
	InstallmentType               string        `json:"produk_jenisangsuran"`
	InstallmentDate               json.Number   `json:"tgl_angsuran"`
	CreditNature2                 string        `json:"sid_sifatkredit2"`
	UsageType                     string        `json:"sid_jenispenggunaan"`
	RepaymentSource               string        `json:"sid_sumberdanapelunasan"`
	CreditGroup                   string        `json:"sid_golongankredit"`
	UsageOrientation              string        `json:"sid_orientasipenggunaan"`
	EconomicSector                string        `json:"sid_sektorekonomi"`
	EconomicSector2               string        `json:"sid_sektorekonomi2"`
	CreditNature                  string        `json:"sid_sifatkredit"`
	GuarantorData                 string        `json:"datapenjamin"`
	SpouseAcknowledgement         string        `json:"mengetahuisuamiistri"`
	BusinessType                  string        `json:"sid_jenisusaha"`
	FlatInterest                  json.Number   `json:"bungaflat"`
	CreditAgreementNumber         string        `json:"noperjanjiankredit"`
	ArrearsPenaltyPercent         json.Number   `json:"persendendatunggakan"`
	Deposit                       json.Number   `json:"titipan"`
	Period                        string        `json:"periode"`
	InterestRate                  json.Number   `json:"produk_sukubunga"`
	InterestRateChange            string        `json:"produk_perubahansukubunga"`
	LastPrincipalInterestPayment  *fincloudDate `json:"tglterakhir_bayarpokokdanbunga"`
	NextPrincipalInterestPayment  *fincloudDate `json:"tglbayarpokokbunga_berikutnya"`
	LastDueDate                   *fincloudDate `json:"tgljtterakhir"`
	NextDueDate                   *fincloudDate `json:"tgljtberikutnya"`
	OutstandingLoan               json.Number   `json:"outstandingpinjaman"`
	PrincipalArrears              json.Number   `json:"tunggakanpokok"`
	Accrue                        json.Number   `json:"accrue"`
	DecimalPoint                  json.Number   `json:"decimalpoint"`
	InterestArrears               json.Number   `json:"tunggakanbunga"`
	ArrearsPenalty                json.Number   `json:"dendatunggakan"`
	DPD                           json.Number   `json:"dpd"`
	KolekBI                       json.Number   `json:"kolekbi"`
	KolekBPR                      json.Number   `json:"kolekbpr"`
	KolekBIUpdate                 string        `json:"updatekolekbi"`
	TotalCollateralValue          json.Number   `json:"totalcollateralvalue"`
	TotalAssetValue               json.Number   `json:"totalassetvalue"`
	CIFNumber                     string        `json:"nocif"`
	CustomerType                  string        `json:"jenisnasabah"`
	CIFOpenDate                   *fincloudDate `json:"tglbukacif"`
	CreditOfficer                 *string       `json:"pejabatkredit"`
	SecondCreditOfficer           *string       `json:"pejabatkreditdua"`
	CreditPurpose                 *string       `json:"tujuankredit"`
	IDCardAddress1                string        `json:"dataalamat_ktp_alamat1"`
	IDCardAddress2                *string       `json:"dataalamat_ktp_alamat2"`
	IDCardRT                      string        `json:"dataalamat_ktp_rt"`
	IDCardRW                      string        `json:"dataalamat_ktp_rw"`
	IDCardVillage                 string        `json:"dataalamat_ktp_kelurahan"`
	IDCardDistrict                string        `json:"dataalamat_ktp_kecamatan"`
	IDCardCity                    string        `json:"dataalamat_ktp_kota"`
	IDCardProvince                string        `json:"dataalamat_ktp_propinsi"`
	IDCardPostalCode              string        `json:"dataalamat_ktp_kodepos"`
	LoanDisbursementSavingAccount string        `json:"norektab_pencairanpinjaman"`
}

// RegisterAutodebitRemoval registers a loan to remove autodebit using the
// caller-selected OPER account.
func (c *Client) RegisterAutodebitRemoval(ctx context.Context, loanAccount, operAccount string) error {
	loanAccount = strings.TrimSpace(loanAccount)
	operAccount = strings.TrimSpace(operAccount)
	if loanAccount == "" {
		return fmt.Errorf("loan account is required")
	}
	if operAccount == "" {
		return fmt.Errorf("OPER account is required")
	}
	if !operAccountPattern.MatchString(operAccount) {
		return fmt.Errorf("invalid OPER account %q", operAccount)
	}

	loan, err := c.inquiryAutodebitRemovalLoan(ctx, loanAccount)
	if err != nil {
		return err
	}
	oper, err := c.inquiryOPERAccount(ctx, operAccount)
	if err != nil {
		return err
	}
	form, err := buildAutodebitRemovalForm(*loan, operAccount, *oper)
	if err != nil {
		return fmt.Errorf("build autodebit removal form: %w", err)
	}
	return c.submitAutodebitRemoval(ctx, loan.AccountNumber, form)
}

func (c *Client) inquiryAutodebitRemovalLoan(ctx context.Context, account string) (*autodebitRemovalLoan, error) {
	resp, err := c.DoRequest(ctx, func() (*http.Request, error) {
		req, err := c.NewRequest(ctx, http.MethodGet, autodebitRemovalInquiryPath, nil)
		if err != nil {
			return nil, err
		}
		query := req.URL.Query()
		query.Set("norekening", account)
		req.URL.RawQuery = query.Encode()
		return req, nil
	})
	if err != nil {
		return nil, fmt.Errorf("autodebit removal loan inquiry %s: %w", account, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("autodebit removal loan inquiry %s: HTTP %d", account, resp.StatusCode)
	}

	var body struct {
		Status string `json:"status"`
		Data   struct {
			Result *autodebitRemovalLoan `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("autodebit removal loan inquiry %s: decode: %w", account, err)
	}
	if body.Status != "ok" {
		return nil, fmt.Errorf("autodebit removal loan inquiry %s: status %q", account, body.Status)
	}
	if body.Data.Result == nil {
		return nil, fmt.Errorf("autodebit removal loan inquiry %s: missing result", account)
	}
	if body.Data.Result.AccountNumber != account {
		return nil, fmt.Errorf("autodebit removal loan inquiry %s: returned account %q", account, body.Data.Result.AccountNumber)
	}
	return body.Data.Result, nil
}

func (c *Client) inquiryOPERAccount(ctx context.Context, account string) (*fincloudapi.SavingBalanceInquiryResponse, error) {
	if c.api == nil {
		return nil, fmt.Errorf("OPER account inquiry %s: %w", account, ErrMissingAPIClient)
	}
	oper, err := c.api.InquirySavingBalance(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("OPER account inquiry %s: %w", account, err)
	}
	if oper == nil || oper.AccountNumber != account {
		return nil, fmt.Errorf("OPER account inquiry %s: returned account %q", account, operAccountNumber(oper))
	}
	if strings.TrimSpace(oper.CustomerName) == "" || strings.TrimSpace(oper.DocumentStatus) == "" || strings.TrimSpace(oper.Currency) == "" {
		return nil, fmt.Errorf("OPER account inquiry %s: missing required fields", account)
	}
	return oper, nil
}

func operAccountNumber(oper *fincloudapi.SavingBalanceInquiryResponse) string {
	if oper == nil {
		return ""
	}
	return oper.AccountNumber
}

func buildAutodebitRemovalForm(loan autodebitRemovalLoan, operAccount string, oper fincloudapi.SavingBalanceInquiryResponse) (url.Values, error) {
	form := url.Values{}
	for key, value := range map[string]string{
		"lokasi":                     loan.Location,
		"id":                         loan.AccountNumber,
		"nopk":                       loan.AgreementNumber,
		"namanasabah":                loan.CustomerName,
		"aliasnama":                  loan.AliasName,
		"statusrekening":             loan.AccountStatus,
		"rec_dibuat_oleh":            loan.CreatedBy,
		"noalt":                      loan.AlternateNumber,
		"produk_jenispinjaman":       loan.LoanType,
		"produkid":                   loan.Product,
		"idproduk":                   loan.ProductID,
		"currency":                   loan.Currency,
		"jangkawaktu":                loan.Term,
		"produk_jenisangsuran":       loan.InstallmentType,
		"sid_sifatkredit2":           loan.CreditNature2,
		"sid_jenispenggunaan":        loan.UsageType,
		"sid_sumberdanapelunasan":    loan.RepaymentSource,
		"sid_golongankredit":         loan.CreditGroup,
		"sid_orientasipenggunaan":    loan.UsageOrientation,
		"sid_sektorekonomi":          loan.EconomicSector,
		"sid_sektorekonomi2":         loan.EconomicSector2,
		"sid_sifatkredit":            loan.CreditNature,
		"datapenjamin":               loan.GuarantorData,
		"mengetahuisuamiistri":       loan.SpouseAcknowledgement,
		"sid_jenisusaha":             loan.BusinessType,
		"noperjanjiankredit":         loan.CreditAgreementNumber,
		"periode":                    loan.Period,
		"produk_perubahansukubunga":  loan.InterestRateChange,
		"updatekolekbi":              loan.KolekBIUpdate,
		"nocif":                      loan.CIFNumber,
		"jenisnasabah":               loan.CustomerType,
		"dataalamat_ktp_alamat1":     loan.IDCardAddress1,
		"dataalamat_ktp_rt":          loan.IDCardRT,
		"dataalamat_ktp_rw":          loan.IDCardRW,
		"dataalamat_ktp_kelurahan":   loan.IDCardVillage,
		"dataalamat_ktp_kecamatan":   loan.IDCardDistrict,
		"dataalamat_ktp_kota":        loan.IDCardCity,
		"dataalamat_ktp_propinsi":    loan.IDCardProvince,
		"dataalamat_ktp_kodepos":     loan.IDCardPostalCode,
		"norektab_pencairanpinjaman": loan.LoanDisbursementSavingAccount,
	} {
		setFormValue(form, key, value)
	}
	for key, value := range map[string]json.Number{
		"plafondlimit":         loan.CreditLimit,
		"jmlpokok_pinjaman":    loan.LoanPrincipal,
		"tgl_angsuran":         loan.InstallmentDate,
		"bungaflat":            loan.FlatInterest,
		"persendendatunggakan": loan.ArrearsPenaltyPercent,
		"titipan":              loan.Deposit,
		"produk_sukubunga":     loan.InterestRate,
		"outstandingpinjaman":  loan.OutstandingLoan,
		"tunggakanpokok":       loan.PrincipalArrears,
		"accrue":               loan.Accrue,
		"decimalpoint":         loan.DecimalPoint,
		"tunggakanbunga":       loan.InterestArrears,
		"dendatunggakan":       loan.ArrearsPenalty,
		"dpd":                  loan.DPD,
		"kolekbi":              loan.KolekBI,
		"kolekbpr":             loan.KolekBPR,
		"totalcollateralvalue": loan.TotalCollateralValue,
		"totalassetvalue":      loan.TotalAssetValue,
	} {
		setFormValue(form, key, value.String())
	}
	for key, value := range map[string]*string{
		"pejabatkredit":          loan.CreditOfficer,
		"pejabatkreditdua":       loan.SecondCreditOfficer,
		"tujuankredit":           loan.CreditPurpose,
		"dataalamat_ktp_alamat2": loan.IDCardAddress2,
	} {
		if value != nil {
			setFormValue(form, key, *value)
		}
	}
	for key, value := range map[string]*fincloudDate{
		"tgl_pencairan":                  loan.DisbursementDate,
		"tglterakhir_bayarpokokdanbunga": loan.LastPrincipalInterestPayment,
		"tglbayarpokokbunga_berikutnya":  loan.NextPrincipalInterestPayment,
		"tgljtterakhir":                  loan.LastDueDate,
		"tgljtberikutnya":                loan.NextDueDate,
		"tglbukacif":                     loan.CIFOpenDate,
	} {
		formatted, err := formatFincloudDate(value)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", key, err)
		}
		setFormValue(form, key, formatted)
	}

	form.Set("status_dokumen", "Diajukan")
	form.Set("norektab_bayarangsuran", operAccount)
	form.Set("tabbayar_namapemilik", oper.CustomerName)
	form.Set("tabbayar_status", oper.DocumentStatus)
	form.Set("tabbayar_currency", oper.Currency)
	return form, nil
}

func setFormValue(form url.Values, key, value string) {
	if strings.TrimSpace(value) != "" {
		form.Set(key, value)
	}
}

func formatFincloudDate(value *fincloudDate) (string, error) {
	if value == nil || strings.TrimSpace(value.Date) == "" {
		return "", nil
	}
	raw := strings.TrimSpace(value.Date)
	if len(raw) < len("2006-01-02") {
		return "", fmt.Errorf("invalid date %q", value.Date)
	}
	date, err := time.Parse("2006-01-02", raw[:10])
	if err != nil {
		return "", fmt.Errorf("invalid date %q: %w", value.Date, err)
	}
	return fmt.Sprintf("%d-%d-%d", date.Year(), date.Month(), date.Day()), nil
}

func (c *Client) submitAutodebitRemoval(ctx context.Context, account string, form url.Values) error {
	encoded := form.Encode()
	resp, err := c.DoRequest(ctx, func() (*http.Request, error) {
		req, err := c.NewRequest(ctx, http.MethodPost, autodebitRemovalSubmitPath, strings.NewReader(encoded))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return req, nil
	})
	if err != nil {
		return fmt.Errorf("autodebit removal submit %s: %w", account, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("autodebit removal submit %s: HTTP %d", account, resp.StatusCode)
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("autodebit removal submit %s: decode: %w", account, err)
	}
	if body.Status != "ok" {
		return fmt.Errorf("autodebit removal submit %s: status %q", account, body.Status)
	}
	return nil
}
