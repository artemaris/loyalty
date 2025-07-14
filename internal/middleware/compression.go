package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// Compression middleware для сжатия ответов
func Compression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, поддерживает ли клиент gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Создаем gzip writer
		gw := gzip.NewWriter(w)
		defer gw.Close()

		// Создаем response writer с gzip
		gzipWriter := &gzipResponseWriter{
			ResponseWriter: w,
			Writer:         gw,
		}

		// Устанавливаем заголовок Content-Encoding
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
