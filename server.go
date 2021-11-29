package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/utils"
)

func test() error {
	list, total, err := loadList("", 0)
	if err != nil {
		return err
	}
	if l := len(list); l == 0 || total == 0 {
		return fmt.Errorf("not expected result. length: %d, total: %d", l, total)
	}

	return nil
}

func run() {
	if *logPath != "" {
		f, err := os.OpenFile(*logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			log.Fatalln("Failed to open log file:", err)
		}
		gin.DefaultWriter = f
		gin.DefaultErrorWriter = f
		log.SetOutput(f)
	}

	router := gin.Default()
	server.Handler = router
	router.TrustedPlatform = "X-Real-IP"

	router.StaticFS("/build", http.Dir(filepath.Join(filepath.Dir(self), "public/build")))
	router.StaticFile("favicon.ico", filepath.Join(filepath.Dir(self), "public/favicon.ico"))
	router.LoadHTMLFiles(
		filepath.Join(filepath.Dir(self), "public/index.html"),
		filepath.Join(filepath.Dir(self), "public/player.html"),
	)

	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	router.GET("/list", func(c *gin.Context) {
		q := c.Query("q")
		p := c.Query("p")
		if p == "" {
			p = "1"
		}

		page, err := strconv.Atoi(p)
		if err != nil {
			c.String(400, "")
			return
		}

		var list []anime
		var total int
		if err := utils.Retry(func() error {
			list, total, err = loadList(q, page)
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

		c.JSON(200, gin.H{"total": total, "list": list})
	})

	router.GET("/play", func(c *gin.Context) {
		c.HTML(200, "player.html", nil)
	})

	router.POST("/play", func(c *gin.Context) {
		var play play
		if err := c.BindJSON(&play); err != nil {
			c.String(400, "")
			return
		}

		var url string
		if err := utils.Retry(func() (err error) {
			url, err = play.loadPlay()
			return
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
