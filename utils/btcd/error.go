package btcd

import "strconv"

// InvalidResponseCodeError is type of error including its response code
type InvalidResponseCodeError struct {
	Code int
}

func (err InvalidResponseCodeError) Error() string {
	return "Responds with invalid response code " + strconv.Itoa(err.Code)
}

// JSONRPCError wraps error response message from btcd
type JSONRPCError struct {
	Code    int
	Message string
}

func (err JSONRPCError) Error() string {
	return "Error response message: " + err.Message + " Code: " + strconv.Itoa(err.Code)
}

// ErrorNoDataReturned indicates no information returned with the given range
const ErrorNoDataReturned string = "No information for the requested range"
