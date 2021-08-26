package main

import (
	"fmt"
	"time"

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

func (p *play) loadPlay() (string, error) {
	id := fmt.Sprintf("%s?playid=%s_%s", p.AID, p.Index, p.EP)
	value, ok := c.Get(id)
	if ok {
		return value.(string), nil
	}

	url, err := p.getPlay()
	if err != nil {
		url, err = p.getPlay2()
		if err != nil {
			return "", err
		}
	}

	c.Set(id, url, time.Hour, nil)

	return url, nil
}
