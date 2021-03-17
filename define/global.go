package define

const (
	RecPerPage = 10
	ClientCheckTicker = 10
	ReqChanSize = 1024
)

//filter kind
const (
	FilterKindMatch = iota + 1
	FilterKindQuery
	FilterKindNumericRange
	FilterKindDateRange
)