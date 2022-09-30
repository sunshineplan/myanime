package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/sunshineplan/chrome"
	"github.com/sunshineplan/gohttp"
)

const prefix = "anime:"

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

func getList(query string, page int) (list []anime, total int, err error) {
	var resp string
	var doc soup.Root
	var roots []soup.Root
	if query == "" {
		resp, err = soup.Get(fmt.Sprintf("%s/update?page=%d", *api, page))
		if err != nil {
			return
		}

		doc = soup.HTMLParse(resp)
		roots = doc.FindAll("li", "class", "anime_icon2")
	} else {
		resp, err = soup.Get(fmt.Sprintf("%s/search?query=%s&page=%d", *api, query, page))
		if err != nil {
			return
		}

		doc = soup.HTMLParse(resp)
		roots = doc.FindAll("div", "class", "cell")
	}

	var href string
	for _, i := range doc.FindAll("a", "class", "pbutton") {
		if i.Text() == "尾页" {
			href = i.Attrs()["href"]
			break
		}
	}

	if strings.Contains(href, "page=") {
		total, err = strconv.Atoi(strings.Split(href, "page=")[1])
		if err != nil {
			return
		}
	} else {
		if query != "" {
			if len(roots) > 0 {
				total = 1
			}
		} else {
			err = fmt.Errorf("unknow error, page: %d", page)
			return
		}
	}

	for _, i := range roots {
		var anime anime

		a := i.Find("a")
		href := a.Attrs()["href"]
		anime.ID = strings.ReplaceAll(href, "/detail/", "")
		anime.URL = *api + href

		img := i.Find("img")
		anime.Name = img.Attrs()["alt"]
		anime.Image = img.Attrs()["src"]

		list = append(list, anime)
	}

	return
}

func getPlayList(u, id string) ([]play, error) {
	resp, err := soup.Get(u)
	if err != nil {
		return nil, err
	}

	doc := soup.HTMLParse(resp)

	index, err := strconv.Atoi(doc.Find("script", "id", "DEF_PLAYINDEX").Text())
	if err != nil {
		return nil, err
	}

	playlist := []play{}
	for _, i := range doc.FindAll("div", "class", "movurl")[index].FindAll("a") {
		href, err := url.Parse(i.Attrs()["href"])
		if err != nil {
			return nil, err
		}
		playid := href.Query().Get("playid")

		playlist = append(playlist, play{
			AID:   id,
			URL:   *api + href.String(),
			Index: strings.Split(playid, "_")[0],
			EP:    strings.Split(playid, "_")[1],
			Title: i.Text(),
		})
	}

	return playlist, nil
}

func (a *anime) getPlayList() error {
	playlist, err := loadPlayList(a.URL, a.ID)
	if err != nil {
		return err
	}

	a.PlayList = playlist

	return nil
}

func (p *play) getPlay() (string, error) {
	s.Get(fmt.Sprintf("%s/play/%s?playid=%s_%s", *api, p.AID, p.Index, p.EP), nil)

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
		fmt.Sprintf("%s/_getplay?aid=%s&playindex=%s&epindex=%s&r=%.f", *api, p.AID, p.Index, p.EP, rand.Float64()),
		gohttp.H{"referer": *api},
	)
	var r struct{ Vurl string }
	if err := resp.JSON(&r); err != nil {
		return "", err
	}

	return parse(resp.Request.URL, r.Vurl)
}

func (p *play) getPlay2() (string, error) {
	c := chrome.Headless(false)
	if _, _, err := c.WithTimeout(10 * time.Second); err != nil {
		return "", err
	}
	defer c.Close()

	if err := c.EnableFetch(func(ev *fetch.EventRequestPaused) bool {
		return ev.ResourceType == network.ResourceTypeDocument ||
			ev.ResourceType == network.ResourceTypeScript ||
			ev.ResourceType == network.ResourceTypeXHR ||
			strings.Contains(ev.Request.URL, "getplay2")
	}); err != nil {
		return "", err
	}

	done := c.ListenEvent(chrome.URLContains("getplay2"), "GET", true)
	if err := chromedp.Run(c, chromedp.Navigate(fmt.Sprintf("%s/play/%s?playid=%s_%s", *api, p.AID, p.Index, p.EP))); err != nil {
		return "", err
	}

	select {
	case <-c.Done():
		return "", c.Err()
	case e := <-done:
		var r struct{ Vurl string }
		if err := json.Unmarshal(e.Bytes, &r); err != nil {
			return "", err
		}

		rurl, _ := url.Parse(e.Request.Request.URL)

		return parse(rurl, r.Vurl)
	}
}

func parse(u *url.URL, vurl string) (string, error) {
	vu, _ := url.QueryUnescape(vurl)
	vURL, err := url.Parse(vu)
	if err != nil {
		return "", err
	}
	if vURL.Host == "" {
		vURL.Scheme = u.Scheme
		vURL.Host = u.Host
	}

	if testM3U8(vURL.String()) {
		resp := gohttp.Get(vURL.String(), nil)
		if resp.Error != nil {
			return "", resp.Error
		}

		m3u8 := resp.String()
		if m3u8 == "" {
			return "", fmt.Errorf("empty m3u8")
		}

		return m3u8, nil
	}

	return prefix + vURL.String(), nil
}

func testM3U8(url string) bool {
	resp := gohttp.Head(url, nil)
	if resp.Error != nil {
		return true
	}
	if resp.StatusCode != 200 {
		return true
	}
	return resp.ContentLength < 3*1024*1024
}
