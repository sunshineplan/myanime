package main

import (
	"flag"
	"net/url"
	"testing"

	"github.com/sunshineplan/utils"
)

func init() {
	flag.StringVar(&api, "api", "", "API")
}

func TestGetList(t *testing.T) {
	list, n, err := getList("", 0)
	if err != nil {
		t.Error(err)
	}
	if len(list) == 0 {
		t.Errorf("expected non-empty list; got empty list")
	}
	if n <= 0 {
		t.Errorf("expected greater than zero; got %d", n)
	}
}

func TestSearch(t *testing.T) {
	list, n, err := getList("go", 0)
	if err != nil {
		t.Error(err)
	}
	if len(list) == 0 {
		t.Errorf("expected non-empty list; got empty list")
	}
	if n <= 0 {
		t.Errorf("expected greater than zero; got %d", n)
	}
}

func TestGetPlayList(t *testing.T) {
	list, err := getPlayList(api+"/detail/20000001", "")
	if err != nil {
		t.Error(err)
	}
	if len(list) == 0 {
		t.Errorf("expected non-empty list; got empty list")
	}
}

func TestGetURL(t *testing.T) {
	var err error
	u, err = url.ParseRequestURI(api)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = ""

	getURL := func(fn func() (string, error)) (res string, err error) {
		err = utils.Retry(
			func() error {
				res, err = fn()
				return err
			}, 3, 5)
		return
	}

	play := play{AID: "20220011", Index: "2", EP: "1"}
	url1, err := getURL(play.getPlay)
	if err != nil {
		t.Fatal(err)
	}
	url2, err := getURL(play.getPlay2)
	if err != nil {
		t.Fatal(err)
	}
	if url1 == "" {
		t.Fatal("getPlay is empty")
	}
	if url1 != url2 {
		t.Fatal("url1 is not same as url2")
	}
}
