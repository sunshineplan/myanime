package main

import (
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
	"github.com/sunshineplan/gohttp"
)

type anime struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Image    string `json:"image"`
	PlayList []play `json:"playlist"`
}

type play struct {
	AID   string `json:"aid"`
	URL   string `json:"url"`
	Index string `json:"index"`
	EP    string `json:"ep"`
	Title string `json:"title"`
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

	doc := soup.HTMLParse(resp)

	index, err := strconv.Atoi(doc.Find("script", "id", "DEF_PLAYINDEX").Text())
	if err != nil {
		return err
	}

	playlist := []play{}
	for _, i := range doc.FindAll("div", "class", "movurl")[index].FindAll("a") {
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

	k2 := fmt.Sprint((t*(t%4096)*3+83215)*(t%4096) + t)
	var t2 string
	for {
		t2 = fmt.Sprint(time.Now().UnixNano() / 1e6)
		if strings.Contains(t2[len(t2)-3:], k2[len(k2)-1:]) {
			break
		}
	}

	s.SetCookie(u, "k2", k2)
	s.SetCookie(u, "t2", t2)
	s.SetCookie(u, "fa_t", fmt.Sprint(time.Now().UnixNano()/1e6))
	s.SetCookie(u, "fa_c", "1")

	resp := s.Get(
		fmt.Sprintf("%s/_getplay?aid=%s&playindex=%s&epindex=%s&r=%.f", api, p.AID, p.Index, p.EP, rand.Float64()),
		gohttp.H{"referer": api},
	)
	var r struct{ Vurl string }
	if err := resp.JSON(&r); err != nil {
		return "", err
	}

	return url.QueryUnescape(r.Vurl)
}
