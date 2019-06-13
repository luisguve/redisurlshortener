package models

import (
	"testing"
	"bytes"
	"reflect"

	"github.com/stretchr/testify/assert"
	"github.com/gomodule/redigo/redis"
)

type testResponse struct {
	OriginalUrl string `json:"original_url"`
	ShortUrl 	string `json:"short_url"`
	ShortUrlId 	string `json:"short_url_id,omitempty"`
	Error 		string `json:"error"`
	Msg 		string `json:"msg"`
}

type testMappingTable struct {
	Input		[]byte
	Expected	string
}

type testBaseConversionTable struct {
	Input 		int64
	Expected 	[]byte
}

type DB_Values struct {
	longUrl 	string
	//dbConn		redis.Conn
}

type SaveOperation struct {
	Input 		DB_Values
	Expected 	testResponse
}

func TestMapIndexes(t *testing.T) {
	assert := assert.New(t)
	tests := []testMappingTable {
		{
			Input:		[]byte{1,2,5},
			Expected:	"abe",
		}, {
			Input:		[]byte{2,1},
			Expected:	"ba",
		}, {
			Input:		[]byte{5,16,47,28,19,57},
			Expected:	"epUBs4",
		}, {
			Input:		[]byte{0},
			Expected:	"",
		}, {
			Input:		[]byte{63},
			Expected:	"ALPHABET overflow",
		}, {
			Input:		[]byte{62},
			Expected:	"9",
		}, 
	}

	var result string

	for _, test := range tests {
		result = mapIndexes(test.Input)
		assert.Equal(result, test.Expected, 
			"inputted: %v, expected: %s, received: %s",
			test.Input, test.Expected, result)
	}
}

func TestBase10ToBase62(t *testing.T) {
	assert := assert.New(t)
	tests := []testBaseConversionTable {
		{
			Input:		int64(125),
			Expected:	[]byte{2,1},
		}, {
			Input:		int64(754),
			Expected:	[]byte{12,10},
		}, {
			Input:		int64(10),
			Expected:	[]byte{10},
		}, {
			Input:		int64(0),
			Expected:	[]byte{0},
		}, {
			Input:		int64(89),
			Expected:	[]byte{1,27},
		}, {
			Input:		int64(265748),
			Expected:	[]byte{1,7,8,16},
		}, {
			Input:		int64(15),
			Expected:	[]byte{15},
		}, {
			Input:		int64(62),
			Expected:	[]byte{1,0},
		}, {
			Input:		int64(63),
			Expected:	[]byte{1,1},
		}, 
	}

	var result []byte
	for _, test := range tests {
		result = base10ToBase62(test.Input)
		assert.Equal(true, bytes.Equal(result, test.Expected),
			"inputted: %v, expected: %v, received: %v",
				test.Input, test.Expected, result)
	}
}

func TestSaveURL(t *testing.T) {
    db := Start("127.0.0.1:6379")

    assert := assert.New(t)

    tests := []SaveOperation {
    	{
    		Input: DB_Values{
	    		longUrl: 	"www.freecodecamp.org",
    		},
    		Expected: testResponse {
				OriginalUrl:	"www.freecodecamp.org",
				ShortUrl:		"localhost:8080/api/shorturl/a",
				ShortUrlId:		"a",
			},
    	}, {
    		Input: DB_Values{
	    		longUrl: 	"www.wikipedia.org",
    		},
    		Expected: testResponse{
				OriginalUrl: 	"www.wikipedia.org",
				ShortUrl:		"localhost:8080/api/shorturl/b",
				ShortUrlId:		"b",
			},
    	}, {
    		Input: DB_Values{
	    		longUrl: 	"www.freecodecamp.org",
    		},
    		Expected: testResponse{
				OriginalUrl: "www.freecodecamp.org",
				   ShortUrl: "localhost:8080/api/shorturl/a",
				 ShortUrlId: "a",
						Msg: "The url www.freecodecamp.org has already been shortened",
    		},
    	},
    }

    var status *Status
    var conn redis.Conn
    for _, test := range tests {
    	func() {
		    conn = db.Get()
		    defer conn.Close()
	    	status = SaveUrl(test.Input.longUrl, conn)
	    	received := testResponse {
	    		OriginalUrl: status.SuccessStatus.OriginalUrl,
	    		ShortUrl: 	 status.SuccessStatus.ShortUrl,
	    		ShortUrlId:  status.SuccessStatus.ShortUrlId,
	    		Msg:		 status.SuccessStatus.Msg,
	    	}
	    	assert.Equal(true, reflect.DeepEqual(received, test.Expected),
	    		"inputted: %v, expected: %v, received: %v",
	    			test.Input, test.Expected, received)
	    }()
    }
}