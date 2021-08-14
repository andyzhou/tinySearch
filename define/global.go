package define

const (
	RecPerPage = 10
	ClientCheckTicker = 3
	ReqChanSize = 1024
)

//query opt kind
const (
	QueryOptKindOfGen = iota
	QueryOptKindOfAgg
	QueryOptKindOfSuggest
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
	FilterKindPhraseQuery
	FilterKindNumericRange
	FilterKindDateRange
	FilterKindSubDocIds
	FilterKindExcludePhraseQuery
)