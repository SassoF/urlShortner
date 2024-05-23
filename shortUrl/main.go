// main
package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"shortUrl/database"
	"strings"
)

type errorResponse struct {
	Error string `json:"error"`
}

type inputMessage struct {
	Url string `json:"longUrl"`
}

type returnMessage struct {
	Url string `json:"shortUrl"`
}

func main() {
	db := database.CreateDatabase()
	defer db.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/favicon.ico", faviconHandler)
	mux.HandleFunc("/api", apiHandler(db))
	mux.HandleFunc("/", urlHandler(db))

	log.Fatal(http.ListenAndServe(":80", mux))

}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../favicon/favicon.ico")
}

func urlHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Write([]byte("Istruzioni scritte in seguito"))
			return
		} else if strings.Count(r.URL.Path, "/") > 1 {
			http.NotFound(w, r)
			return
		}

		url, err := database.GetUrl(db, database.QUERY_GET_LONG_URL, strings.TrimLeft(r.URL.Path, "/")+"=")
		if err != nil {
			log.Println("GetUrl:", err)
			sendErrorRequest(w, "Internal error")
			return
		} else if url != "" {
			http.Redirect(w, r, url, http.StatusMovedPermanently)
			return
		}

		http.NotFound(w, r)

	}
}

func apiHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var inputMessage inputMessage
		if err := json.NewDecoder(r.Body).Decode(&inputMessage); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if inputMessage.Url == "" {
			sendErrorRequest(w, "Parametro url non presente")
			return
		}

		var err error
		var rMessage returnMessage
		rMessage.Url, err = database.InsertUrl(db, &inputMessage.Url)

		if err != nil {
			if message, ok := err.(*database.MessageError); ok {
				sendErrorRequest(w, message.Error())
			} else {
				log.Println("InserUrl:", err)
				sendErrorRequest(w, "Internal error")
			}
			return
		}

		rMessage.Url = "127.0.0.1/" + strings.TrimRight(rMessage.Url, "=")

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(rMessage); err != nil {
			sendErrorRequest(w, "Errore interno")
			return
		}
	}
}

func sendErrorRequest(w http.ResponseWriter, errorMessage string) {
	w.WriteHeader(http.StatusInternalServerError)
	errResponse := errorResponse{Error: errorMessage}
	if err := json.NewEncoder(w).Encode(errResponse); err != nil {
		log.Println("Errore nell'invio della risposta di errore", err)
	}
}
