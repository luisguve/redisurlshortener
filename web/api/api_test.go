package api

import(
	"fmt"
	"reflect"
	"encoding/json"
	"net/url"
	"net/http"
	"net/http/httptest"
	"testing"
	"strings"

	"github.com/villegasl/urlshortener.redis/models"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type testResponse struct {
	OriginalUrl string `json:"original_url"`
	ShortUrl 	string `json:"short_url"`
	ShortUrlId 	string `json:"short_url_id,omitempty"`
	Error 		string `json:"error"`
	Msg 		string `json:"msg"`
}

func TestAPI(t *testing.T) {
	assert := assert.New(t)
	testsNewShortUrl := []testResponse {
		{
			OriginalUrl: "www.freecodecamp.org",
			   ShortUrl: "localhost:8080/api/shorturl/a",
			 ShortUrlId: "a",
		}, {
			OriginalUrl: "www.wikipedia.org",
			   ShortUrl: "localhost:8080/api/shorturl/b",
			 ShortUrlId: "b",
		}, {
			OriginalUrl: "www.freecodecamp.org",
			   ShortUrl: "localhost:8080/api/shorturl/a",
			 ShortUrlId: "a",
					Msg: "The url www.freecodecamp.org has already been shortened",
		}, 
	}
	testsRedirectByShortUrl := []testResponse {
		{
			OriginalUrl: "www.freecodecamp.org",
			   ShortUrl: "localhost:8080/api/shorturl/a",
			 ShortUrlId: "a",
		}, {
			OriginalUrl: "www.wikipedia.org",
			   ShortUrl: "localhost:8080/api/shorturl/b",
			 ShortUrlId: "b",
		}, {
				 Error: "unable to find the long URL",
			ShortUrlId: "777",
		},
	}

	db := models.Start("127.0.0.1:6379")

	// Variable status will hold the status code received from every request
	var status int

	// Testing the function: NewShortUrl
	for _, expected := range testsNewShortUrl {
		handler := NewShortURL(db)
		form := url.Values{}
		form.Add("url", expected.OriginalUrl)
		r := httptest.NewRequest("POST", "/api/shorturl/new", 
			strings.NewReader(form.Encode()))
		r.Form = form
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		// Check the response body is what we expect.
		status = w.Code
		assert.Equal(status, http.StatusOK, 
			"handler returned wrong status code: got %v want %v", 
			status, http.StatusOK)
		var received testResponse
		err := json.Unmarshal(w.Body.Bytes(), &received)
		if err != nil {
			fmt.Println("error unmarshaling:", err.Error())
			fmt.Println("value received:", w.Body.String())
			return
	    }
	    assert.Equal(true, reflect.DeepEqual(expected, received),
	    	"handler returned unexpected body: got %v want %v",
			received, expected)
	}

	// Testing the function: RedirectByShortUrl
	// This test must follow the test to the function NewShortUrl
	for _, expected := range testsRedirectByShortUrl {
		path := "/api/shorturl/" + expected.ShortUrlId
		r := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		/*Why instantiate a gorilla/mux instance when we could have 
		simply done RedirectByShortURL(DB_Handler).ServeHTTP(w,r) ? 
		This is because in the handler implementation, 
		we had to retrieve the url param. 
		Not instantiating mux would mean we wouldn’t be able 
		to fetch the url parameter.*/

		router.Handle("/api/shorturl/{url:[a-zA-Z0-9]{1,11}}",
			RedirectByShortURL(db)).Methods(http.MethodGet)

		router.ServeHTTP(w, r)
		status = w.Code
		assert.Equal(status, http.StatusOK,
			"handler returned wrong status code: got %v want %v",
			status, http.StatusOK)

		var received testResponse
		err := json.Unmarshal(w.Body.Bytes(), &received)
		if err != nil {
			fmt.Println("error unmarshaling:", err.Error())
			fmt.Println("value received:", w.Body.String())
			return
	    }
	    assert.Equal(true, reflect.DeepEqual(expected, received),
	    	"handler returned unexpected body.\n Got: %#v\n\n Want: %#v\n\n",
			received, expected)
	}
}