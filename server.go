package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/utils/retry"
)

var (
	urlParse         = url.Parse
	urlQueryEscape   = url.QueryEscape
	urlQueryUnescape = url.QueryUnescape
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
		if err := retry.Do(func() error {
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
				if e := retry.Do(a.getPlayList, 3, 1); e != nil {
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

		var res string
		if err := retry.Do(func() (err error) {
			res, err = play.loadPlay()
			return
		}, 3, 3); err != nil {
			log.Print(err)
			c.String(500, "")
			return
		}

		if strings.HasPrefix(res, prefix) {
			url := strings.TrimPrefix(res, prefix)
			if testM3U8(url) {
				c.String(200, fmt.Sprint("/m3u8?ref=", urlQueryEscape(url)))
			} else {
				c.String(200, urlQueryEscape(url))
			}
		} else {
			c.String(200, fmt.Sprintf("/m3u8/%s/%s/%s", play.AID, play.Index, play.EP))
		}
	})

	router.GET("/m3u8", func(c *gin.Context) {
		url := c.Query("ref")
		url, err := urlQueryUnescape(url)
		if err != nil {
			log.Print(err)
			c.String(404, "")
			return
		}

		res, err := loadM3U8(url)
		if err == nil {
			c.Data(200, "application/vnd.apple.mpegurl", []byte(res))
		} else {
			log.Print(err)
			c.String(404, "")
		}
	})

	router.GET("/m3u8/:aid/:index/:ep", func(c *gin.Context) {
		play := play{AID: c.Param("aid"), Index: c.Param("index"), EP: c.Param("ep")}
		res, err := play.loadPlay()
		if err == nil {
			if strings.HasPrefix(res, prefix) {
				c.String(200, strings.TrimPrefix(res, prefix))
			} else {
				c.Data(200, "application/vnd.apple.mpegurl", []byte(res))
			}
		} else {
			log.Print(err)
			c.String(404, "")
		}
	})

	router.NoRoute(func(c *gin.Context) {
		c.Redirect(302, "/")
	})

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
