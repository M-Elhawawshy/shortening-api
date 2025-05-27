package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (app *application) redirectHandler(w http.ResponseWriter, r *http.Request) {
	urlHash := strings.TrimPrefix(r.URL.Path, "/")
	if urlHash == "" {
		app.clientError(w, r, fmt.Errorf("empty link"), http.StatusBadRequest)
		return
	}

	link, err := app.cache.Get(r.Context(), urlHash).Result()
	if err == nil {
		http.Redirect(w, r, link, http.StatusFound)
		return
	}

	dbLink, err := app.queries.GetLink(r.Context(), urlHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			app.clientError(w, r, err, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}
	_, err = app.cache.Set(r.Context(), urlHash, dbLink.Link.String, time.Hour*24*3).Result()
	if err != nil {
		app.logger.Error("redis failed to cache the link redirect request")
	}

	http.Redirect(w, r, dbLink.Link.String, http.StatusFound)
}
