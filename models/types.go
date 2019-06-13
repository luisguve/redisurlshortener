package models

import(
	"github.com/gomodule/redigo/redis"
)

type DBHandler struct {
	*redis.Pool
}

type SuccessResponse struct {
	OriginalUrl string `json:"original_url"`
	ShortUrl 	string `json:"short_url"`
	ShortUrlId	string `json:"short_url_id"`
	Msg 		string `json:"msg,omitempty"`
}

type FailureResponse struct {
	Error 		string `json:"error"`
	ShortUrlId	string `json:"short_url_id,omitempty"`
}

type Status struct {
	Error 			error
	SuccessStatus 	*SuccessResponse
	FailureStatus 	*FailureResponse
}

type TestDB struct {
    *DBHandler
}

// The range of values within the slice of base 62-encoded digits 
// goes from 1 to 62.
// But the ALPHABET array is 0-indexed. Its range goes from 0 to 61.
// In order to keep the compatibility between this array and the
// digits within the base 62-encoded slice, the first value in the array
// is an empty string, and now the range of the valid values in the array
// i.e. [a-z-A-Z-0-9] goes from 1 to 62.
var ALPHABET = []string{
	"", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", 
	"l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", 
	"y", "z","A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", 
	"L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", 
	"Y", "Z", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
}