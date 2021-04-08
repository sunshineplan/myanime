package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/sunshineplan/gohttp"
	"github.com/vharitonsky/iniflags"
)

var api, self string
var u *url.URL

var s = gohttp.NewSession()

func init() {
	var err error
	self, err = os.Executable()
	if err != nil {
		log.Fatalln("Failed to get self path:", err)
	}

	gohttp.SetAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.114 Safari/537.36")
}

func main() {
	flag.StringVar(&api, "api", "https://www.agefans.net", "API")
	iniflags.SetConfigFile(filepath.Join(filepath.Dir(self), "config.ini"))
	iniflags.SetAllowMissingConfigFile(true)
	iniflags.SetAllowUnknownFlags(true)
	iniflags.Parse()

	var err error
	u, err = url.ParseRequestURI(api)
	if err != nil {
		log.Fatal(err)
	}
	u.Path = ""

	list, err := getList()
	if err != nil {
		log.Fatal(err)
	}

	if err := list[0].getPlayList(); err != nil {
		log.Fatal(err)
	}

	r, err := list[0].PlayList[0].getURL()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(list[0].PlayList[0], r)
}
