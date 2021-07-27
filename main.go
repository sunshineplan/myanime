package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sunshineplan/gohttp"
	"github.com/sunshineplan/service"
	"github.com/sunshineplan/utils"
	"github.com/sunshineplan/utils/httpsvr"
	"github.com/vharitonsky/iniflags"
)

var api, self string
var logPath *string
var server httpsvr.Server
var u *url.URL

var s = gohttp.NewSession()
var svc = service.Service{
	Name:     "MyAnime",
	Desc:     "Instance to serve My Anime",
	Exec:     run,
	TestExec: test,
	Options: service.Options{
		Dependencies: []string{"After=network.target"},
		Environment:  map[string]string{"GIN_MODE": "release"},
	},
}

func init() {
	var err error
	self, err = os.Executable()
	if err != nil {
		log.Fatalln("Failed to get self path:", err)
	}

	gohttp.SetAgent(utils.UserAgent(
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36",
	))

	rand.Seed(time.Now().UnixNano())
}

func usage(errmsg string) {
	fmt.Fprintf(os.Stderr,
		`%s
usage: %s <command>
       where <command> is one of install, remove, start, stop.
`, errmsg, os.Args[0])
	os.Exit(2)
}

func main() {
	flag.StringVar(&api, "api", "", "API")
	flag.StringVar(&server.Unix, "unix", "", "UNIX-domain Socket")
	flag.StringVar(&server.Host, "host", "0.0.0.0", "Server Host")
	flag.StringVar(&server.Port, "port", "12345", "Server Port")
	flag.StringVar(&svc.Options.UpdateURL, "update", "", "Update URL")
	exclude := flag.String("exclude", "", "Exclude Files")
	//logPath = flag.String("log", joinPath(dir(self), "access.log"), "Log Path")
	logPath = flag.String("log", "", "Log Path")
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

	svc.Options.ExcludeFiles = strings.Split(*exclude, ",")

	if service.IsWindowsService() {
		svc.Run(false)
		return
	}

	switch flag.NArg() {
	case 0:
		run()
	case 1:
		switch flag.Arg(0) {
		case "run", "debug":
			svc.Run(true)
		case "test":
			err = svc.Test()
		case "install":
			err = svc.Install()
		case "remove":
			err = svc.Remove()
		case "start":
			err = svc.Start()
		case "stop":
			err = svc.Stop()
		case "restart":
			err = svc.Restart()
		case "update":
			err = svc.Update()
		default:
			usage(fmt.Sprintf("Unknown argument: %s", flag.Arg(0)))
		}
	default:
		usage(fmt.Sprintf("Unknown arguments: %s", strings.Join(flag.Args(), " ")))
	}
	if err != nil {
		log.Fatalf("Failed to %s: %v", flag.Arg(0), err)
	}
}
