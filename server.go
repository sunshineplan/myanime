package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/utils"
)

func run() {
	router := gin.Default()
	server.Handler = router

	if *logPath != "" {
		f, err := os.OpenFile(*logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			log.Fatalln("Failed to open log file:", err)
		}
		gin.DefaultWriter = f
		gin.DefaultErrorWriter = f
		log.SetOutput(f)
	}

	router.StaticFS("/build", http.Dir(filepath.Join(filepath.Dir(self), "public/build")))
	router.StaticFile("favicon.ico", filepath.Join(filepath.Dir(self), "public/favicon.ico"))
	router.LoadHTMLFiles(filepath.Join(filepath.Dir(self), "public/index.html"))

	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	router.GET("/list", func(c *gin.Context) {
		q := c.Query("q")

		var list []anime
		var err error
		if err := utils.Retry(func() error {
			list, err = getList(q)
			return err
		}, 3, 2); err != nil {
			log.Print(err)
			c.String(500, "")
			return
		}

		var wg sync.WaitGroup
		wg.Add(len(list))
		for i := range list {
			go func(a *anime) {
				defer wg.Done()
				if e := utils.Retry(a.getPlayList, 3, 1); e != nil {
					err = e
					log.Print(err)
				}
			}(&list[i])
		}
		wg.Wait()

		if err != nil {
			c.String(500, "")
			return
		}

		c.JSON(200, list)
	})

	router.POST("/play", func(c *gin.Context) {
		var play play
		if err := c.BindJSON(&play); err != nil {
			c.String(400, "")
			return
		}

		var url string
		var err error
		if err := utils.Retry(func() error {
			url, err = play.getURL()
			return err
		}, 3, 3); err != nil {
			log.Print(err)
			c.String(500, "")
			return
		}

		c.String(200, url)
	})

	router.NoRoute(func(c *gin.Context) {
		c.Redirect(302, "/")
	})

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
