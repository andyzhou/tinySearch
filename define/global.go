package define

const (
	RecPerPage = 10
	ClientCheckTicker = 10
	ReqChanSize = 1024
)

//query kind
const (
	QueryKindOfTerm = iota + 1
	QueryKindOfMatchQuery
	QueryKindOfPhrase
	QueryKindOfPrefix
	QueryKindOfNumericRange
)

//filter kind
const (
	FilterKindMatch = iota + 1
	FilterKindMatchRange
	FilterKindQuery
	FilterKindNumericRange
	FilterKindDateRange
	FilterKindSubDocIds
)