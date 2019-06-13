package api

import(
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/villegasl/urlshortener.redis/models"

	log "github.com/sirupsen/logrus"
	"github.com/gorilla/mux"
)

func RedirectByShortURL(DB_Handler *models.DBHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn := DB_Handler.Get()
		log.Infoln("conn should be done now")
		defer log.Infoln("conn should be closed now")
		defer conn.Close()
		vars := mux.Vars(r)
		shortURL := vars["url"]

		log.Infoln("\nRequested short url:",shortURL)

		status := models.GetUrl(shortURL, conn)

		if status.Error != nil {
			log.Errorln("could not find the corresponding long url:", 
				status.Error)
			if status.Error == models.ErrNoUrl {
				respondWithJSON(w, http.StatusOK, status.FailureStatus)
				return
			}
			respondWithJSON(w, http.StatusInternalServerError, status.FailureStatus)
			return
		}

		log.Infoln("Corresponding long url:",status.SuccessStatus.OriginalUrl)
		respondWithJSON(w, http.StatusOK, status.SuccessStatus)
	})
}

func NewShortURL(DB_Handler *models.DBHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn := DB_Handler.Get()
		log.Infoln("conn should be done now")
		defer log.Infoln("conn should be closed now")
		defer conn.Close()
		originalUrl := r.FormValue("url")

		// Please uncomment this url validation when good internet speed be available
		/*_, err := http.Get(originalUrl)
		if err != nil {
			fmt.Println("Error: Invalid URL:", err.Error())
			jsonRes, err := json.Marshal(models.FailureResponse { Error: "Invalid URL" })
			if err != nil {
				fmt.Println("Error while trying to marshal the JSON response:", err)
				return
			}
			w.Write(jsonRes)
			return
		}*/

		status := models.SaveUrl(originalUrl, conn)

		if status.Error != nil {
			fmt.Println("could not update the database:",status.Error)
			respondWithError(w, http.StatusInternalServerError, 
	"could not update the database: " + status.FailureStatus.Error)
			return
		}
		// At this point the long url was shortened successfully

		// Respond to the client with the appropiate JSON
		respondWithJSON(w, http.StatusOK, status.SuccessStatus)
	})
}

// RespondWithError is called on an error to return info regarding error
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// Called for responses to encode and send json data
func respondWithJSON(w http.ResponseWriter, code int, res interface{}) {
	//encode response to json
	response, err := json.Marshal(res)
	if err != nil {
		fmt.Println("marshal returned error:",err.Error())
	}

	// set headers and write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
