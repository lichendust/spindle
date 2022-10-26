//go:build !debug

package main

import "hash/fnv"

func new_hash(input string) uint32 {
	if input == "" {
		return 0
	}
	hash := fnv.New32a()
	hash.Write([]byte(input))
	return hash.Sum32()
}