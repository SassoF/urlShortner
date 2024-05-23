package database

import (
	"database/sql"
	"log"
	"shortUrl/urlHandler"
	"strings"

	"github.com/go-sql-driver/mysql"
)

type MessageError struct {
	message string
}

func (e *MessageError) Error() string {
	return e.message
}

const (
	QUERY_ADD_URL       = "INSERT INTO link(longUrl, shortUrl) VALUES (?, ?)"
	QUERY_GET_LONG_URL  = "SELECT longUrl FROM link WHERE shortUrl = ?"
	QUERY_GET_SHORT_URL = "SELECT shortUrl FROM link WHERE longUrl = ?"
)

func CreateDatabase() *sql.DB {

	db, err := sql.Open("mysql", "username:password@tcp(localhost:3306)/databaseUrl")
	if err != nil {
		panic(err)
	}
	// Crea una tabella "link" se non esiste
	createTable := `
		CREATE TABLE IF NOT EXISTS link (
			longUrl TEXT NOT NULL UNIQUE,
			shortUrl TEXT NOT NULL UNIQUE
		);
	`

	_, err = db.Exec(createTable)
	if err != nil {
		panic(err)
	}
	return db
}

func GetUrl(db *sql.DB, QUERY string, url string) (string, error) {
	var urlDB string
	err := db.QueryRow(QUERY, url).Scan(&urlDB)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return urlDB, nil
}

func InsertUrl(db *sql.DB, longUrl *string) (string, error) {

	//check if longUrl is valid
	if !urlHandler.IsValidUrl(longUrl) {
		return "", &MessageError{"Url non valido"}
	}

	//check if shortUrl of longUrl exists
	//if return "" doesn't exist
	if shortUrl, err := GetUrl(db, QUERY_GET_SHORT_URL, *longUrl); shortUrl == "" {
		//if the error isn't nil, there's another error
		if err != nil {
			log.Println("GetUrl:", err)
			return "", err
		} //if the error is nil the code continues
		//shortUrl exists in the database
	} else {
		return shortUrl, nil
	}

	//for 3 times try to insert longUrl
	for _ = range 3 {
		//generate shortUrl and check for errors
		var shortUrl string
		{
			var err error
			if shortUrl, err = urlHandler.GenerateShortUrl(); err != nil {
				log.Println("generateShortUrl:", err)
				continue
			}
		}

		//insert longUrl and shortUrl
		result, err := db.Exec(QUERY_ADD_URL, longUrl, shortUrl)
		if err != nil {
			if mysqlErr, ok := err.(*mysql.MySQLError); ok {
				if mysqlErr.Number == 1062 {
					if strings.Contains(mysqlErr.Message, "shortUrl") {
						log.Println("shortUrl alredy exist", shortUrl)
						continue
					} else if strings.Contains(mysqlErr.Message, "longUrl") {
						if shortUrl, err := GetUrl(db, QUERY_GET_SHORT_URL, *longUrl); shortUrl == "" {
							//if the error isn't nil, there's another error
							if err != nil {
								log.Println("GetUrl:", err)
								return "", err
							} //if the error is nil the code continues
							//shortUrl exists in the database
						} else {
							return shortUrl, nil
						}

					}
				}
				//if error isn't 1062 (unique) log error
				log.Println("exec", err)
			}
		} else {
			//checks if only one row has been modified
			rows, err := result.RowsAffected()
			if err != nil || rows != 1 {
				log.Println("rows:", err)
				return "", &MessageError{"Internal error"}
			}
		}
		//no error
		return shortUrl, nil
	}

	//after 3 attempts it was not entered after errors
	return "", &MessageError{"Internal error"}

}
