package main

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
)

// MiddlewareEncrypt - middleware for encrypting request body with crypto key
func (App *Application) MiddlewareEncrypt(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		var result []byte
		if strings.Contains(r.Header.Get("X-Encrypted"), "rsa") {
			_, err := buf.ReadFrom(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			result, err = data.DecryptData(App.CryptoKey, buf.Bytes())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				App.Logger.Errorln("Error while decrypting data:", err)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(result))
			r.Header.Del("X-Encrypted")
		}

		next(w, r)
	}
}

// MiddlewareUnpack - middleware for unpacking request body into zip archive
func (App *Application) MiddlewareUnpack(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				App.Logger.Errorln("Error during unpacking the request: ", err)
				return
			}
			defer gz.Close()

			_, err = io.Copy(&buf, gz)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			r.Body = io.NopCloser(&buf)
			r.Header.Del("Content-Encoding")
		}

		next(w, r)
	}
}

// MiddlewareZipper - middleware for packing response into zip archive.
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

// MiddlewareLogger - function for logging information about processing request.
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

// MiddlewareTrustedIP - function for checking, if client IP-address is trusted
func (App *Application) MiddlewareTrustedIP(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		agentIP := r.Header.Get("X-Real-IP")
		if agentIP != "" {
			_, cidr, err := net.ParseCIDR(App.TrustedSubnet)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				App.Logger.Errorln("Error during parsing CIDR")
				return
			}
			 
			trustedIP := cidr.Contains(net.ParseIP(agentIP))
			if !trustedIP {
				http.Error(w, "Untrusted IP-adress: access denied", http.StatusForbidden)
				App.Logger.Errorln("Untrusted IP-adress: access denied")
				return
			}
		}
		next(w, r)
	}
}

// MiddlewareHash - function for checking data integrity of request body.
func (App *Application) MiddlewareHash(next http.HandlerFunc) http.HandlerFunc {
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

		body, err := io.ReadAll(r.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			App.Logger.Errorln("Error during reading request body")
			return
		}

		// Replace the body with a new reader after reading from the original
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		sign := r.Header.Get("HashSHA256")
		if sign != "" {
			signDecode, err := hex.DecodeString(sign)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				App.Logger.Errorln("Error during HashSHA256 decoding")
			}

			h := hmac.New(sha256.New, []byte(App.SecretKey))
			h.Write(body)
			signCheck := h.Sum(nil)

			if !(hmac.Equal(signDecode, signCheck)) {
				http.Error(w, "Error while checking HashSHA256 of the request", http.StatusBadRequest)
				App.Logger.Errorln("HashSHA256 is incorrect")
				return
			}
		}

		next(&zlw, r)
	}
}

// MiddlewareChain - function for processing chain of middlewares.
func (App *Application) MiddlewareChain(h http.HandlerFunc, m ...data.Middleware) http.HandlerFunc {
	for _, wrap := range m {
		h = wrap(h)
	}

	return h
}
