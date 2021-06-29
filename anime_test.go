package main

import (
	"flag"
	"fmt"
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

	var res string
	expected := "//www.iqiyi.com/v_19rrok4nt0.html"
	if err := utils.Retry(
		func() error {
			res, err = (&play{AID: "20000001", Index: "2", EP: "1"}).getURL()
			if err != nil {
				t.Fatal(err)
			}

			if res == expected {
				return nil
			}
			return fmt.Errorf("not match")
		}, 3, 5); err != nil {
		t.Errorf("expected %q; got %q", expected, res)
	}
}
