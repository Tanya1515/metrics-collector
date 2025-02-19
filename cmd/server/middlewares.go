package main

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"net/http"
	"strings"
	"time"

	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
)

func (App *Application) MiddlewareZipper(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responseData := &ResponseData{
			status: 0,
			size:   0,
		}

		zlw := LoggingZipperResponseWriter{
			w,
			w,
			responseData,
		}
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				App.Logger.Errorln("Error during comressing")
			}

			defer gz.Close()
			w.Header().Set("Content-Encoding", "gzip")
			zlw = LoggingZipperResponseWriter{
				w,
				gz,
				responseData,
			}
		}

		next(&zlw, r)

	}
}

func (App *Application) MiddlewareLogger(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uri := r.RequestURI
		method := r.Method

		responseData := &ResponseData{
			status: 0,
			size:   0,
		}

		zlw := LoggingZipperResponseWriter{
			w,
			w,
			responseData,
		}

		start := time.Now()

		h.ServeHTTP(&zlw, r)
		duration := time.Since(start)

		App.Logger.Infoln(
			"URI", uri,
			"Method", method,
			"Duration", duration,
			"ResponseStatus", responseData.status,
			"ResponseSize", responseData.size,
		)

	}
}

func (App *Application) MiddlewareHash(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var reqBody []byte

		_, err := r.Body.Read(reqBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			App.Logger.Errorln("Error during reading request body")
			return
		}

		sign := r.Header.Get("HashSHA256")

		h := hmac.New(sha256.New, []byte(secretKeyHash))
		h.Write(reqBody)
		signCheck := h.Sum(nil)

		if !(hmac.Equal([]byte(sign), signCheck)) {
			http.Error(w, "Error while checking HashSHA256 of the request", http.StatusBadRequest)
			App.Logger.Errorln("HashSHA256 is incorrect")
			return
		}
	}
}

// ... - variardic parameter, that can get any amount of parameters of type data.Middleware
func (App *Application) MiddlewareChain(h http.HandlerFunc, m ...data.Middleware) http.HandlerFunc {
	for _, wrap := range m {
		h = wrap(h)
	}

	return h
}
