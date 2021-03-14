package define

const (
	RecPerPage = 10
	ClientCheckTicker = 10
	ReqChanSize = 128
)

//filter kind
const (
	FilterKindMatch = iota + 1
	FilterKindQuery
	FilterKindNumericRange
	FilterKindDateRange
)