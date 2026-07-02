package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"mytab/server/internal/storage"
)

const (
	defaultAddr    = "127.0.0.1:8080"
	defaultDataDir = "data"
)

type apiError struct {
	Error string `json:"error"`
}

func main() {
	addr := env("ADDR", defaultAddr)
	dataDir := env("DATA_DIR", defaultDataDir)

	store, err := storage.New(dataDir)
	if err != nil {
		log.Fatalf("create storage: %v", err)
	}

	router := gin.Default()
	router.MaxMultipartMemory = storage.MaxFiles * storage.MaxFileBytes
	router.Use(cors())

	router.GET("/healthz", func(c *gin.Context) {
		c.Header("X-MyTab", "true")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api")
	{
		api.GET("/projects", func(c *gin.Context) {
			projects, err := store.List()
			if err != nil {
				respondError(c, http.StatusInternalServerError, err)
				return
			}
			c.JSON(http.StatusOK, projects)
		})

		api.POST("/projects", func(c *gin.Context) {
			form, err := c.MultipartForm()
			if err != nil {
				respondError(c, http.StatusBadRequest, err)
				return
			}

			manifest, err := store.Create(c.PostForm("name"), form.File["files"], form.Value["paths"])
			if err != nil {
				respondStorageError(c, err)
				return
			}
			c.JSON(http.StatusCreated, manifest)
		})

		api.GET("/projects/:id", func(c *gin.Context) {
			manifest, err := store.Get(c.Param("id"))
			if err != nil {
				respondStorageError(c, err)
				return
			}
			c.JSON(http.StatusOK, manifest)
		})

		api.GET("/projects/:id/files/*file", func(c *gin.Context) {
			serveStoredFile(c, store, false)
		})

		api.GET("/projects/:id/preview/*file", func(c *gin.Context) {
			serveStoredFile(c, store, true)
		})
	}

	mountFrontend(router)

	log.Printf("listening on %s", displayURL(addr))
	if err := router.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func displayURL(addr string) string {
	switch {
	case strings.HasPrefix(addr, ":"):
		return "http://localhost" + addr
	case strings.HasPrefix(addr, "0.0.0.0:"):
		return "http://localhost:" + strings.TrimPrefix(addr, "0.0.0.0:")
	case strings.HasPrefix(addr, "[::]:"):
		return "http://localhost:" + strings.TrimPrefix(addr, "[::]:")
	default:
		return "http://" + addr
	}
}

func serveStoredFile(c *gin.Context, store *storage.Store, preview bool) {
	filePath, kind, err := store.ResolveFile(c.Param("id"), c.Param("file"))
	if err != nil {
		respondStorageError(c, err)
		return
	}

	c.Header("X-Content-Type-Options", "nosniff")
	if ct := contentType(kind); ct != "" {
		c.Header("Content-Type", ct)
	}
	if preview {
		c.Header("Content-Security-Policy", strings.Join([]string{
			"default-src 'none'",
			"script-src 'self' 'unsafe-inline'",
			"style-src 'self' 'unsafe-inline'",
			"img-src 'self' data: blob:",
			"font-src 'self' data:",
			"media-src 'self' data: blob:",
			"connect-src 'none'",
			"object-src 'none'",
			"base-uri 'self'",
			"form-action 'none'",
		}, "; "))
	}

	file, err := os.Open(filePath)
	if err != nil {
		respondStorageError(c, err)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		respondStorageError(c, err)
		return
	}

	http.ServeContent(c.Writer, c.Request, filepath.Base(filePath), stat.ModTime(), file)
}

func respondStorageError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, storage.ErrNotFound), errors.Is(err, storage.ErrFileNotFound):
		respondError(c, http.StatusNotFound, err)
	case errors.Is(err, storage.ErrInvalidUpload):
		respondError(c, http.StatusBadRequest, err)
	default:
		respondError(c, http.StatusInternalServerError, err)
	}
}

func respondError(c *gin.Context, status int, err error) {
	c.JSON(status, apiError{Error: err.Error()})
}

func contentType(kind string) string {
	switch kind {
	case "html":
		return "text/html; charset=utf-8"
	case "css":
		return "text/css; charset=utf-8"
	case "javascript":
		return "text/javascript; charset=utf-8"
	case "json":
		return "application/json; charset=utf-8"
	case "text":
		return "text/plain; charset=utf-8"
	case "image", "font", "media":
		return ""
	default:
		return "application/octet-stream"
	}
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func mountFrontend(router *gin.Engine) {
	dist := env("WEB_DIST", filepath.Clean(filepath.Join("..", "web", "dist")))
	index := filepath.Join(dist, "index.html")
	if _, err := os.Stat(index); err != nil {
		return
	}

	router.Static("/assets", filepath.Join(dist, "assets"))
	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, apiError{Error: "not found"})
			return
		}
		c.File(index)
	})
}

func env(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
