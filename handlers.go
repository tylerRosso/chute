// SPDX-License-Identifier: ISC

package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type FileInfo struct {
	Name string
	Size string
}

func setupRouter(uploadDir string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex(uploadDir))
	mux.HandleFunc("/upload", handleUpload(uploadDir))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))
	return mux
}

func generateDirETag(entries []os.DirEntry) (string, []FileInfo) {
	h := sha256.New()
	h.Write(templateHash[:])

	var fileList []FileInfo
	for _, f := range entries {
		if !f.IsDir() {
			info, err := f.Info()
			if err == nil {
				fmt.Fprintf(h, "%s:%d:%d;", f.Name(), info.Size(), info.ModTime().UnixNano())
				fileList = append(fileList, FileInfo{
					Name: f.Name(),
					Size: formatSize(info.Size()),
				})
			}
		}
	}
	etag := fmt.Sprintf(`"%x"`, h.Sum(nil))
	return etag, fileList
}

func matchETag(ifNoneMatch, etag string) bool {
	if ifNoneMatch == "" {
		return false
	}
	if ifNoneMatch == "*" {
		return true
	}
	cleanETag := strings.Trim(strings.TrimPrefix(etag, "W/"), `"`)
	for _, item := range strings.Split(ifNoneMatch, ",") {
		cleanItem := strings.Trim(strings.TrimPrefix(strings.TrimSpace(item), "W/"), `"`)
		if cleanItem == cleanETag {
			return true
		}
	}
	return false
}

func handleIndex(uploadDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		entries, err := os.ReadDir(uploadDir)
		if err != nil {
			entries = nil
		}

		etag, fileList := generateDirETag(entries)

		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("ETag", etag)

		if (r.Method == http.MethodGet || r.Method == http.MethodHead) && matchETag(r.Header.Get("If-None-Match"), etag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		indexTmpl.Execute(w, map[string]any{"Files": fileList})
	}
}

func handleUpload(uploadDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		reader, err := r.MultipartReader()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if part.FileName() == "" {
				continue
			}

			filename := filepath.Base(part.FileName())
			dstPath := filepath.Join(uploadDir, filename)

			dst, err := os.Create(dstPath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if _, err := io.Copy(dst, part); err != nil {
				dst.Close()
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			dst.Close()
		}

		w.WriteHeader(http.StatusOK)
	}
}
