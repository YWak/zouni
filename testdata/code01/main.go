package main

import "example.com/lib"

func indirect(v int) int {
	return v
}

func main() {
	print(indirect(lib.Add(1, 2)))
}
