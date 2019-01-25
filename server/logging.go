package main

import (
	"bufio"
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

type loggedResponseWriter struct {
	StatusCode int
	http.ResponseWriter
}

func (l *loggedResponseWriter) WriteHeader(status int) {
	l.StatusCode = status
	l.ResponseWriter.WriteHeader(status)
}

func (l *loggedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := l.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("Wrapped response does not support hijacking")
	}
	return hj.Hijack()
}

func logRequestsMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		logger := &loggedResponseWriter{
			ResponseWriter: resp,
			StatusCode:     200,
		}

		start := time.Now()
		var (
			body []byte
			err  error
		)
		contentType, ok := req.Header["Content-Type"]
		noPrintBody := false
		if ok && (len(contentType) > 0) {
			noPrintBody = strings.HasPrefix(contentType[0], "multipart/form-data")
		}
		if noPrintBody {
			body = []byte("File Upload")
		} else {
			body, err = httputil.DumpRequest(req, true)
			if err != nil {
				body = []byte("N/A")
			}
		}
		handler.ServeHTTP(logger, req)
		elapsed := roundElapsedDuration(time.Since(start), time.Microsecond*10)

		bodyStr := string(body)
		dblNewlineIndex := strings.Index(string(body), "\r\n\r\n")
		if dblNewlineIndex != -1 {
			bodyStr = bodyStr[dblNewlineIndex+4:]
		}
		bodyStr = strings.Replace(bodyStr, "\n", "\n\t\t", -1)

		log.Printf(
			" -- [%v] [%13s] %20s %v: %v\r\n\t%s",
			logger.StatusCode,
			elapsed,
			"",
			req.Method,
			req.RequestURI,
			bodyStr,
		)
	})
}

func roundElapsedDuration(d, r time.Duration) time.Duration {
	if r <= 0 {
		return d
	}
	neg := d < 0
	if neg {
		d = -d
	}
	if m := d % r; m+m < r {
		d = d - m
	} else {
		d = d + r - m
	}
	if neg {
		return -d
	}
	return d
}
