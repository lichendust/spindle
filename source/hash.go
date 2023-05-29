/*
	Spindle
	A static site generator
	Copyright (C) 2022-2023 Harley Denham

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

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