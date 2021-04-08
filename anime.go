package main

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
	"github.com/sunshineplan/gohttp"
)

type anime struct {
	ID       string
	Name     string
	URL      string
	Image    string
	PlayList []play
}

type play struct {
	AID   string
	URL   string
	Index string
	EP    string
	Title string
}

func getList() ([]anime, error) {
	resp, err := soup.Get(fmt.Sprintf("%s/update", api))
	if err != nil {
		return nil, err
	}

	var result []anime
	doc := soup.HTMLParse(resp)
	for _, i := range doc.FindAll("li", "class", "anime_icon2") {
		var anime anime

		a := i.Find("a")
		href := a.Attrs()["href"]
		anime.ID = strings.ReplaceAll(href, "/detail/", "")
		anime.URL = api + href

		img := i.Find("img")
		anime.Name = img.Attrs()["alt"]
		anime.Image = img.Attrs()["src"]

		result = append(result, anime)
	}

	return result, nil
}

func (a *anime) getPlayList() error {
	resp, err := soup.Get(a.URL)
	if err != nil {
		return err
	}

	var movurl soup.Root
	doc := soup.HTMLParse(resp)
	for _, i := range doc.FindAll("div", "class", "movurl") {
		if i.Attrs()["style"] == "display:block" {
			movurl = i
			break
		}
	}
	if movurl.Pointer == nil {
		return errors.New("failed to get movurl")
	}

	var playlist []play
	for _, i := range movurl.FindAll("a") {
		href, err := url.Parse(i.Attrs()["href"])
		if err != nil {
			return err
		}
		playid := href.Query().Get("playid")

		playlist = append(playlist, play{
			AID:   a.ID,
			URL:   api + href.String(),
			Index: strings.Split(playid, "_")[0],
			EP:    strings.Split(playid, "_")[1],
			Title: i.Text(),
		})
	}
	a.PlayList = playlist

	return nil
}

func (p *play) getURL() (string, error) {
	s.Get(fmt.Sprintf("%s/play/%s?playid=%s_%s", api, p.AID, p.Index, p.EP), nil)

	var t1 float64
	for _, i := range s.Cookies(u) {
		if i.Name == "t1" {
			t1, _ = strconv.ParseFloat(i.Value, 64)
			break
		}
	}
	t := int(math.Round(t1/1000)) >> 5

	s.SetCookie(u, "t2", fmt.Sprint(time.Now().UnixNano()/1e6))
	s.SetCookie(u, "k2", fmt.Sprint((t*(t%4096)*3+83215)*(t%4096)+t))
	s.SetCookie(u, "fa_t", fmt.Sprint(time.Now().UnixNano()/1e6))
	s.SetCookie(u, "fa_c", "1")

	resp := s.Get(
		fmt.Sprintf("%s/_getplay?aid=%s&playindex=%s&epindex=%s", api, p.AID, p.Index, p.EP),
		gohttp.H{"referer": api},
	)
	var r struct{ Vurl string }
	if err := resp.JSON(&r); err != nil {
		return "", err
	}

	return url.QueryUnescape(r.Vurl)
}
