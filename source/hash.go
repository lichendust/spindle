package main

import "hash/fnv"

var hash_store = make(map[uint32]string, 256)

func new_hash(input string) uint32 {
	if input == "" {
		return 0
	}

	hash := fnv.New32a()
	hash.Write([]byte(input))
	x := hash.Sum32()

	if _, ok := hash_store[x]; !ok {
		hash_store[x] = input
	}

	return x
}

func get_hash(n uint32) string {
	return hash_store[n]
}