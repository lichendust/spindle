package main

var cache_plate map[string]*markup
var cache_rtext map[string]string

func init() {
	cache_plate = make(map[string]*markup, 4)
	cache_rtext = make(map[string]string,  4)
}

func expire_cache_plate() {
	cache_plate = make(map[string]*markup, len(cache_plate))
}

func expire_cache_rtext() {
	cache_rtext = make(map[string]string,  len(cache_rtext))
}