package define

//internal define
const (
	SuggestTopMin          = 50
	SuggestTopMax          = 200
	RecPerPage             = 10
	ClientCheckTicker      = 5
	ReqChanSize            = 1024
	DataPathDefault        = "./private"
	InterSuggestChanSize   = 1024
	CustomTokenizerOfJieBa = "jieba"

	InterDefaultGroup     = "__group__"
	InterSuggestIndexPara = "__suggester_%v"
)

//default value
const (
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
	QueryKindOfMatchAll = iota + 1
	QueryKindOfTerm
	QueryKindOfMatchQuery
	QueryKindOfPhrase
	QueryKindOfMatchPhraseQuery
	QueryKindOfPrefix
	QueryKindOfGeoDistance
	QueryKindOfConjunctionQuery
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
	FilterKindBoolean
	FilterKindTermsQuery
)