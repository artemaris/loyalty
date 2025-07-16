package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func Compression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gw := gzip.NewWriter(w)
		defer gw.Close()

		gzipWriter := &gzipResponseWriter{
			ResponseWriter: w,
			Writer:         gw,
		}

		w.Header().Set("Content-Encoding", "gzip")

		next.ServeHTTP(gzipWriter, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (g *gzipResponseWriter) Write(data []byte) (int, error) {
	return g.Writer.Write(data)
}
