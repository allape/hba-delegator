package main

import (
	_ "embed"
	"github.com/allape/gocrud"
	"github.com/allape/gogger"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const IndexHTMLName = "index.html"

//go:embed index.html
var IndexHTMLTemplate string

var l = gogger.New("main")

type History struct {
	gocrud.Base
	URL string `json:"url"`
}

func main() {
	err := gogger.InitFromEnv()
	if err != nil {
		l.Error().Fatalln(err)
	}

	dl := logger.New(gogger.New("db").Debug(), logger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      logger.Info,
		Colorful:      true,
	})
	db, err := gorm.Open(sqlite.Open(DatabaseFilename), &gorm.Config{
		Logger: dl,
	})
	if err != nil {
		l.Error().Fatalf("failed to create database: %v", err)
	}

	err = db.AutoMigrate(&History{})
	if err != nil {
		l.Error().Fatalf("failed to auto migrate database: %v", err)
	}

	engine := gin.Default()

	render := template.Must(template.New(IndexHTMLName).Delims("{{", "}}").Funcs(map[string]any{}).Parse(IndexHTMLTemplate))

	renderIndex := func(context *gin.Context, params map[string]any) {
		if params == nil {
			params = map[string]any{}
		}

		if _, ok := params["histories"]; !ok {
			var histories []History
			db.Find(&histories)
			params["histories"] = histories
		}

		context.HTML(http.StatusOK, IndexHTMLName, params)
	}

	engine.SetHTMLTemplate(render)
	engine.GET("/", func(context *gin.Context) {
		renderIndex(context, nil)
	})

	engine.POST("/jump", func(context *gin.Context) {
		dst := context.PostForm("dst")

		if dst == "" {
			renderIndex(context, map[string]any{
				"error": "dst is required",
			})
			return
		}

		u, err := url.Parse(dst)
		if err != nil {
			renderIndex(context, map[string]any{
				"error": err.Error(),
			})
			return
		}

		username := context.PostForm("username")
		password := context.PostForm("password")

		if username == "" {
			renderIndex(context, map[string]any{
				"error": "username is required",
			})
			return
		}

		u.User = url.UserPassword(username, password)

		history := History{
			URL: dst,
		}

		err = db.Model(&history).First(&history).Error
		if err != nil {
			db.Create(&history)
		} else {
			db.Model(&history).Update("updated_at", time.Now())
		}

		context.Redirect(http.StatusTemporaryRedirect, u.String())
	})

	engine.GET("/favicon.ico", func(context *gin.Context) {
		context.AbortWithStatus(http.StatusNotFound)
	})

	engine.NoRoute(func(context *gin.Context) {
		context.Redirect(http.StatusMovedPermanently, "/")
	})

	go func() {
		err = engine.Run(HttpAddress)
		if err != nil {
			l.Error().Fatalln("failed to start server", err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigs
	l.Info().Println("Exiting with", sig)
}
