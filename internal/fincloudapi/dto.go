package fincloudapi

type CIFInquiryRequest struct {
	AccountNumber    string `json:"accountNumber"`
	NationalIdNumber string `json:"nationalIdNo"`
	CIFNumber        string `json:"cifNo"`
}

type CreateCIFRequest struct {
	Name             string `json:"name"`
	MobilePhoneNo    string `json:"mobilePhoneNo"`
	NationalIDNo     string `json:"nationalIdNo"`
	BirthDate        string `json:"birthDate"`
	Gender           int64  `json:"gender"`
	Email            string `json:"email"`
	TaxIDNo          string `json:"taxIdNo"`
	Nationality      string `json:"nationality"`
	BirthPlace       string `json:"birthPlace"`
	MotherMaidenName string `json:"motherMaidenName"`
	Religion         string `json:"religion"`
	Province         string `json:"province"`
	City             string `json:"city"`
	PostalCode       string `json:"postalCode"`
	SubDistrict      string `json:"subDistrict"`
	Village          string `json:"village"`
	Address          string `json:"address"`
	MaritalStatus    string `json:"maritalStatus"`
	LastEducation    string `json:"lastEducation"`
	Job              string `json:"job"`
	IncomeSource     string `json:"incomeSource"`
	IncomeRange      string `json:"incomeRange"`
	JobCompanyName   string `json:"jobCompanyName"`
	JobType          string `json:"jobType"`
	JobAddress       string `json:"jobAddress"`
	JobPostalCode    string `json:"jobPostalCode"`
	JobPhone         string `json:"jobPhone"`
	User             string `json:"user"`
	BranchCode       string `json:"branchCode"`
	CityDati2        string `json:"cityDati2"`
	ReferenceName    string `json:"referenceName"`
	Address2         string `json:"address2"`
	Neighbourhood    string `json:"neighbourhood"`
	Hamlet           string `json:"hamlet"`
}

type SavingStatementInquiryRequest struct {
	AccountNumber string `json:"accountNo"`
	StartDate     string `json:"startDate"`
	EndDate       string `json:"endDate"`
	RowPerPage    uint   `json:"rowPerPage"`
	Index         uint   `json:"index"`
}

type GlToGlRequest struct {
	ReferenceNumber string `json:"referenceNumber"`
	TrxType         string `json:"trxType"`
	TermType        string `json:"termType"`
	TermID          string `json:"termId"`
	ReceiptNumber   string `json:"receiptNumber"`
	SourceAccount   string `json:"sourceAccount"`
	DebitAccount    string `json:"debitAccount"`
	CreditAccount   string `json:"creditAccount"`
	Amount          string `json:"amount"`
	Fee             string `json:"fee"`
	CreditFee       string `json:"creditFee"`
	BranchCode      string `json:"branchCode"`
	DebitNarrative  string `json:"debitNarrative"`
	CreditNarrative string `json:"creditNarrative"`
	CustomerID      string `json:"customerId"`
	DateTime        string `json:"dateTime"`
	Description     string `json:"description"`
	DebitFee        string `json:"debitFee"`
	DestAccount     string `json:"destAccount"`
	Currency        string `json:"currency"`
	SrcAccType      string `json:"srcAccType"`
	TotalBill       string `json:"totalBill"`
	Type            string `json:"type"`
}

type GlToSavingRequest struct {
	Amount          string `json:"amount"`
	CustomerID      string `json:"customerId"`
	DateTime        string `json:"dateTime"`
	Description     string `json:"description"`
	Fee             string `json:"fee"`
	DebitFee        string `json:"debitFee"`
	CreditFee       string `json:"creditFee"`
	ReferenceNumber string `json:"referenceNumber"`
	DestAccount     string `json:"destAccount"`
	ReceiptNumber   string `json:"receiptNumber"`
	Currency        string `json:"currency"`
	BranchCode      string `json:"branchCode"`
	SrcAccType      string `json:"srcAccType"`
	TermID          string `json:"termId"`
	TermType        string `json:"termType"`
	TotalBill       string `json:"totalBill"`
	Type            string `json:"type"`
}

type LoanTerminationRequest struct {
	TrxReference   string `json:"trxReference"`
	AccountNumber  string `json:"accountNumber"`
	AltNumber      string `json:"altNumber"`
	PrincipalPaid  int64  `json:"principalPaid"`
	InterestPaid   int64  `json:"interestPaid"`
	PenaltyPaid    int64  `json:"penaltyPaid"`
	PrincipalWaive int64  `json:"principalWaive"`
	InterestWaive  int64  `json:"interestWaive"`
	Description    string `json:"description"`
	BranchCode     string `json:"branchCode"`
}

type LoanRepaymentRequest struct {
	TrxReference  string `json:"trxReference"`
	AccountNumber string `json:"accountNumber"`
	AltNumber     string `json:"altNumber"`
	PaymentAmount int64  `json:"paymentAmount"`
	Description   string `json:"description"`
	BranchCode    string `json:"branchCode"`
}

type LoanInquiryRequest struct {
	AccountNumber string `json:"accountNumber"`
	AltNumber     string `json:"altNumber"`
}

type LoanDisbursementRequest struct {
	IdProduct               string `json:"idProduct"`
	Description             string `json:"description"`
	HashCode                string `json:"hashCode"`
	LoanNumber              string `json:"loanNumber"`
	PkNumber                string `json:"pkNumber"`
	OrderId                 string `json:"orderId"`
	AltCif                  string `json:"altCif"`
	BranchCode              string `json:"branchCode"`
	FullName                string `json:"fullName"`
	IdNumber                string `json:"idNumber"`
	IdCardExpiredDate       string `json:"idCardExpiredDate"`
	SpouseIdCardNumber      string `json:"spouseIdCardNumber"`
	SpouseIdCardExpiredDate string `json:"spouseIdCardExpiredDate"`
	Npwp                    string `json:"npwp"`
	BirthPlace              string `json:"birthPlace"`
	BirthDate               string `json:"birthDate"`
	Mmn                     string `json:"mmn"`
	LastEducation           string `json:"lastEducation"`
	Religion                string `json:"religion"`
	Gender                  string `json:"gender"`
	MaritalStatus           string `json:"maritalStatus"`
	Address                 string `json:"address"`
	Village                 string `json:"village"`
	District                string `json:"district"`
	CityDati2               string `json:"cityDati2"`
	Province                string `json:"province"`
	ZipCode                 string `json:"zipCode"`
	PhoneNo                 string `json:"phoneNo"`
	MobileNo                string `json:"mobileNo"`
	HomeStatus              string `json:"homeStatus"`
	LiveSince               string `json:"liveSince"`
	IdCardRtRw              string `json:"idCardRtRw"`
	SourceOfFunds           string `json:"sourceOfFunds"`
	JobId                   string `json:"jobId"`
	CorpName                string `json:"corpName"`
	CorpAddress             string `json:"corpAddress"`
	CorpCity                string `json:"corpCity"`
	CorpProvince            string `json:"corpProvince"`
	CorpZipCode             string `json:"corpZipCode"`
	EconomyCode             string `json:"economyCode"`
	EmployeeNo              string `json:"employeeNo"`
	JobTitle                string `json:"jobTitle"`
	WorkSince               string `json:"workSince"`
	GrossIncome             string `json:"grossIncome"`
	IncomeRange             string `json:"incomeRange"`
	Expenses                string `json:"expenses"`
	CreditCardBankName      string `json:"creditCardBankName"`
	CreditCardPrincipal     string `json:"creditCardPrincipal"`
	BankName                string `json:"bankName"`
	CreditPrincipal         string `json:"creditPrincipal"`
	PrincipalTotal          string `json:"principalTotal"`
	Tenor                   string `json:"tenor"`
	Interest                string `json:"interest"`
	Score                   string `json:"score"`
	Email                   string `json:"email"`
	BorrowerKtp             string `json:"borrowerKtp"`
	SelfiePhoto             string `json:"selfiePhoto"`
	OtherPhoto              string `json:"otherPhoto"`
	PhotoNpwp               string `json:"photoNpwp"`
	PhotoKk                 string `json:"photoKk"`
	PhotoAkta               string `json:"photoAkta"`
	PhotoSkdu               string `json:"photoSkdu"`
	SalarySlipPicture       string `json:"salarySlipPicture"`
	AccountBusinessPhoto    string `json:"accountBusinessPhoto"`
	CollateralPhoto         string `json:"collateralPhoto"`
	DocCollateralPhoto      string `json:"docCollateralPhoto"`
	FrontPhoto              string `json:"frontPhoto"`
	BehindPhoto             string `json:"behindPhoto"`
	BorrowerStatus          string `json:"borrowerStatus"`
	PefindoScore            string `json:"pefindoScore"`
	FaceRecognition         string `json:"faceRecognition"`
	VoiceRecognition        string `json:"voiceRecognition"`
	UserBehavior            string `json:"userBehavior"`
	BlackListChecking       string `json:"blackListChecking"`
	OutletName              string `json:"outletName"`
	OutletAddress           string `json:"outletAddress"`
	ContactPerson           string `json:"contactPerson"`
	EmployeeGrade           string `json:"employeeGrade"`
	Majelis                 string `json:"majelis"`
	BranchName              string `json:"branchName"`
	CorpPhone               string `json:"corpPhone"`
	LifeInsuranceCompany    string `json:"lifeInsuranceCompany"`
	PremiAjk                string `json:"premiAjk"`
	CreditInsuranceCompany  string `json:"creditInsuranceCompany"`
	PremiCreditInsurance    string `json:"premiCreditInsurance"`
	BusinessType            string `json:"businessType"`
	LoanDescr               string `json:"loanDescr"`
	SlikTerms               string `json:"slikTerms"`
	OtherObligations        string `json:"otherObligations"`
	CollateralForm          string `json:"collateralForm"`
	CollateralType          string `json:"collateralType"`
	CollateralMerk          string `json:"collateralMerk"`
	CollateralDocNumber     string `json:"collateralDocNumber"`
	CollateralColor         string `json:"collateralColor"`
	CollateralPoliceNumber  string `json:"collateralPoliceNumber"`
	CollateralMacNumber     string `json:"collateralMacNumber"`
	InsuranceName           string `json:"insuranceName"`
	InsuranceValue          string `json:"insuranceValue"`
	DisburseType            string `json:"disburseType"`
	CollateralAge           string `json:"collateralAge"`
	CollateralValue         string `json:"collateralValue"`
	Ltv                     string `json:"ltv"`
	Marketing               string `json:"marketing"`
	NickName                string `json:"nickName"`
	RelationId              string `json:"relationId"`
	AreaCode                string `json:"areaCode"`
	CategoryId              string `json:"categoryId"`
	RelationBankId          string `json:"relationBankId"`
	CreditAttributeId       string `json:"creditAttributeId"`
	TypeOfUseId             string `json:"typeOfUseId"`
	OrientationOfUseId      string `json:"orientationOfUseId"`
	CreditCategoryId        string `json:"creditCategoryId"`
	SaForLoanRepayment      string `json:"saForLoanRepayment"`
}

type InquiryETOrRepaymentRequest struct {
	AccountNumber string `json:"accountNumber"`
	AltNumber     string `json:"altNumber"`
	TrxReference  string `json:"trxReference"`
}

type TransferOverbookingRequest struct {
	TrxReference  string `json:"trxReference"`
	TrxType       string `json:"trxType"`
	TermType      string `json:"termType"`
	TermID        string `json:"termId"`
	ReceiptNumber string `json:"receiptNumber"`
	BranchCode    string `json:"branchCode"`
	SrcAccount    string `json:"srcAccount"`
	DestAccount   string `json:"destAccount"`
	Amount        string `json:"amount"`
	DebitFee      int64  `json:"debitFee"`
	CreditFee     int64  `json:"creditFee"`
	Description   string `json:"description"`
}

type CreateSavingRequest struct {
	CifNumber      string `json:"cifNumber"`
	ProductCode    string `json:"productCode"`
	Alias          string `json:"alias"`
	OpeningPurpose string `json:"openingPurpose"`
	AltNumber      string `json:"altNumber"`
	BranchCode     string `json:"branchCode"`
	FacilityValue  string `json:"facilityValue"`
}

type BlockSavingRequest struct {
	AccountNumber   string `json:"accountNumber"`
	AlternateNumber string `json:"alternateNumber"`
	BlockAmount     int64  `json:"blockAmount"`
	EndDate         string `json:"endDate"`
	BlockReason     string `json:"blockReason"`
	Description     string `json:"description"`
	BranchCode      string `json:"branchCode"`
}

type UnblockSavingRequest struct {
	AccountNumber   string `json:"accountNumber"`
	AlternateNumber string `json:"alternateNumber"`
	UnblockReason   string `json:"unblockReason"`
	Description     string `json:"description"`
	BranchCode      string `json:"branchCode"`
}
