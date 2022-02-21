package main

import (
	"fmt"
	"time"

	"github.com/sunshineplan/hlsdl"
	"github.com/sunshineplan/utils/cache"
)

type key struct {
	query string
	page  int
}

type result struct {
	list  []anime
	total int
}

var c = cache.New(true)

func loadList(query string, page int) (list []anime, total int, err error) {
	value, ok := c.Get(key{query: query, page: page})
	if ok {
		list = value.(result).list
		total = value.(result).total
		return
	}

	list, total, err = getList(query, page)
	if err != nil {
		return
	}

	c.Set(
		key{query: query, page: page},
		result{list: list, total: total},
		10*time.Minute,
		nil,
	)

	return
}

func loadPlayList(url, id string) (playlist []play, err error) {
	value, ok := c.Get(id)
	if ok {
		playlist = value.([]play)
		return
	}

	playlist, err = getPlayList(url, id)
	if err != nil {
		return
	}

	c.Set(id, playlist, time.Hour, nil)

	return
}

func loadM3U8(url string) (string, error) {
	value, ok := c.Get(url)
	if ok {
		return value.(string), nil
	}

	u, err := urlParse(url)
	if err != nil {
		return "", err
	}

	_, m3u8, err := hlsdl.FetchM3U8MediaPlaylist(u, false)
	if err != nil {
		return "", err
	}

	c.Set(url, m3u8.String(), time.Hour, nil)

	return m3u8.String(), nil
}

func (p *play) loadPlay() (string, error) {
	id := fmt.Sprintf("%s?playid=%s_%s", p.AID, p.Index, p.EP)
	value, ok := c.Get(id)
	if ok {
		return value.(string), nil
	}

	m3u8, err := p.getPlay()
	if err != nil {
		if *chrome {
			m3u8, err = p.getPlay2()
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	c.Set(id, m3u8, time.Hour, nil)

	return m3u8, nil
}
