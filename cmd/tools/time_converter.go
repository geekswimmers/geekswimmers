package main

import "geekswimmers/utils"

func main() {
	miliseconds, err := utils.MillisecondsFromText("01:24.99")
	if err != nil {
		panic(err)
	}
	println(miliseconds)
}
