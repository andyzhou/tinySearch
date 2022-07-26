package define

//internal define
const (
	RecPerPage = 10
	ClientCheckTicker = 3
	ReqChanSize = 1024
	DataPathDefault = "."
)

//default value
const (
	SearchDirDefault = "./private"
	SearchRpcPortDefault = 6060
	SearchDictFileDefault = "./private/dict.txt"
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
	QueryKindOfMatchAll
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
	FilterKindPrefix
)