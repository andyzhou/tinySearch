package define

//internal define
const (
	RecPerPage = 10
	ClientCheckTicker = 3
	ReqChanSize = 1024
	DataPathDefault = "."
	InterSuggestIndexPara = "__suggester_%v"
	InterSuggestChanSize = 1024
	CustomTokenizerOfJieBa = "jieba"
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
	QueryKindOfTerm = iota + 1
	QueryKindOfMatchQuery
	QueryKindOfPhrase
	QueryKindOfPrefix
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