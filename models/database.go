package models

import (
	"fmt"
    "time"
	"errors"

    "github.com/gomodule/redigo/redis"
)

var ErrNoUrl = errors.New("unable to find the long URL")

func GetUrl(shortUrlId string, db redis.Conn) *Status {
    result, err := redis.StringMap(db.Do("HGETALL", "url_data:"+shortUrlId))
    switch {
    case len(result) == 0:
        fmt.Println("(empty list or set)")
        return &Status {
            Error:          ErrNoUrl,
            FailureStatus:  &FailureResponse {
                Error:      ErrNoUrl.Error(),
                ShortUrlId: shortUrlId,
            },
        }
    case err != nil:
        fmt.Println("error redis.StringMap():",err.Error())
        return &Status{
            Error:          err,
            FailureStatus:  &FailureResponse{ Error: err.Error() },
        }
    default:
        return &Status {
            Error:          nil,
            SuccessStatus:  &SuccessResponse{
                OriginalUrl: result["original"],
                   ShortUrl: result["shortened"],
                 ShortUrlId: shortUrlId,
            },
            FailureStatus:  nil,
        }
    }
}

func SaveUrl(longUrl string, db redis.Conn) *Status {
    // Get the short_url_id to check wheher the URL is new.
    shortUrlId, err := redis.String(db.Do("HGET", "long_url:"+longUrl, "short_url_id"))
    switch {
    case err == redis.ErrNil:
        // At this point we are sure the url has not been shortened before.
        // Check if this is the first entry in redis
        urlCounter, err := redis.Int64(db.Do("GET", "url_counter"))
        switch {
        case err == redis.ErrNil:
            // If so, just set the url counter to 1.
            fmt.Printf("FIRST ENTRY ON THE DATABASE: %s\n\n", longUrl)
            _, err = db.Do("SET", "url_counter", "1")
            if err != nil {
                fmt.Println("error db.Do(SET url_counter 1):",err.Error())
                return &Status{
                    Error:          err,
                    FailureStatus:  &FailureResponse{ Error: err.Error() },
                }
            }
        case err != nil:
            fmt.Println("error db.Do(GET url_counter):",err.Error())
            return &Status {
                Error:          err,
                FailureStatus:  &FailureResponse{ Error: err.Error() },
            }
        }
        urlCounter++
        // Convert the base 10 int64 to base 62.
        // This function returns a []byte.
        IDByteSlice := base10ToBase62(urlCounter)
        // The []byte will be parsed to the string
        // representation of the base 62 number.
        shortUrlId = mapIndexes(IDByteSlice)
        // The new short url.
        shortUrl := "localhost:8080/api/shorturl/" + shortUrlId

        // It must perform three operations in a single transaction:
        // shorten and save the url data - save the url record - set the url counter
        
        // Start transaction
        err = db.Send("MULTI")
        if err != nil {
            fmt.Println("error db.Send(MULTI):",err.Error())
            return &Status {
                Error:          err,
                FailureStatus:  &FailureResponse{ Error: err.Error() },
            }
        }
        // Queue up the transaction's commands
        if err = db.Send("SET", "url_counter", urlCounter); err != nil {
            fmt.Println("error db.Send(SET url_counter):",err.Error())
            return &Status {
                Error:          err,
                FailureStatus:  &FailureResponse{ Error: err.Error() },
            }
        }
        if err = db.Send("HMSET", "url_data:"+shortUrlId, 
                "original", longUrl, "shortened", shortUrl); err != nil {
            fmt.Println("error db.Send(HMSET url_data):", err.Error())
            return &Status {
                Error:          err,
                FailureStatus:  &FailureResponse{ Error: err.Error() },
            }
        }
        if err = db.Send("HSET", "long_url:"+longUrl,
                "short_url_id", shortUrlId); err != nil {
            fmt.Println("error db.Send(HSET long_url):", err.Error())
            return &Status {
                Error:          err,
                FailureStatus:  &FailureResponse{ Error: err.Error() },
            }
        }
        // Finish transaction
        _, err = db.Do("EXEC")
        if err != nil{
            fmt.Println("error db.Do(EXEC):", err.Error())
            return &Status {
                Error:          err,
                FailureStatus:  &FailureResponse{ Error: err.Error() },
            }
        }
        fmt.Println("New URL saved:",longUrl)
        fmt.Println("New short URL:", shortUrl)
        fmt.Printf("New short URL id: %s\n\n", shortUrlId)
        return &Status {
            Error:          nil,
            SuccessStatus:  &SuccessResponse {
             OriginalUrl:   longUrl,
                ShortUrl:   shortUrl,
              ShortUrlId:   shortUrlId,
            },
            FailureStatus:  nil,
        }
    case err != nil:
        fmt.Println("error db.Do(HGET long_url):",err.Error())
        return &Status{
            Error:          err,
            FailureStatus:  &FailureResponse{ Error: err.Error() },
        }
    }
    // At this point we know the URL has already been shortened before.
    fmt.Println("URL", longUrl, "has already been shortened")
    // Get the rest of the data.
    result, err := redis.StringMap(db.Do("HGETALL", "url_data:"+shortUrlId))
    if err != nil {
        fmt.Println("error db.Do():",err.Error())
        return &Status{
            Error:          err,
            FailureStatus:  &FailureResponse{ Error: err.Error() },
        }
    }
    fmt.Println("short URL:", result["shortened"])
    fmt.Printf("short URL id: %s\n\n", shortUrlId)
    return &Status {
        Error:           nil,
        SuccessStatus: &SuccessResponse{
                OriginalUrl: result["original"],
                ShortUrl:    result["shortened"],
                ShortUrlId:  shortUrlId,
                Msg:         fmt.Sprintf("The url %s has already been shortened", longUrl),
        },
        FailureStatus:      nil,
    }
}

func Start(addr string) *DBHandler {
	pool := newPool(addr)

	return &DBHandler{pool}
}

func newPool(addr string) *redis.Pool {
    return &redis.Pool{
        MaxIdle: 10,
        IdleTimeout: 5 * time.Second,
        Dial: func () (redis.Conn, error) {
            c, err := redis.Dial("tcp", addr)
            if err != nil {
                panic(err)
            }
            return c, err
        },
    }
}