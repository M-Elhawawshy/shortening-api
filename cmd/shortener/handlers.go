package main

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jxskiss/base62"
	"net/http"
	"net/url"
	"shortening-api/internal/database"
	"shortening-api/internal/helpers"
)

type LinkSubmissionForm struct {
	Link string `form:"link" json:"link"`
}

func (app *application) shortenerHandler(w http.ResponseWriter, r *http.Request) {
	var linkForm LinkSubmissionForm

	if err := helpers.ParseForm(r, &linkForm); err != nil {
		app.clientError(w, r, err, http.StatusBadRequest)
		return
	}

	link := linkForm.Link
	if link == "" {
		app.clientError(w, r, fmt.Errorf("empty link"), http.StatusBadRequest)
		return
	}

	URL, err := url.Parse(link)
	if err != nil || URL.Scheme == "" || URL.Host == "" {
		app.clientError(w, r, fmt.Errorf("empty link"), http.StatusBadRequest)
		return
	}
	hash := sha256.New()
	_, err = hash.Write([]byte(URL.String()))
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	userID := r.Header.Get("X-User-ID")
	app.logger.Debug("userID: " + userID)
	useruuid, err := uuid.Parse(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	hashedURLBytes := hash.Sum(nil)
	var encodedStr string
	var inserted bool = false

	for i := 7; i < 10; i++ {
		if inserted {
			break
		}

		encodedStr = base62.EncodeToString(hashedURLBytes[:i])
		_, err = app.queries.GetLink(r.Context(), encodedStr)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// insert the link into db
				_, err = app.queries.InsertLink(r.Context(), database.InsertLinkParams{
					Hash:   encodedStr,
					UserID: useruuid,
					Link: pgtype.Text{
						String: link,
						Valid:  true,
					},
				})
				if err != nil {
					app.serverError(w, r, err)
					return
				}
				inserted = true
			} else {
				// db failed for some reason
				app.serverError(w, r, err)
				return
			}
		}
	}

	if inserted {
		_, _ = w.Write([]byte(encodedStr))
	} else {
		// some sorcery with the links
		app.clientError(w, r, fmt.Errorf("tried 3 times with hash but failed"), http.StatusBadRequest)
	}
}
