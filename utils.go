package main

import "os"

// Contains check if the following array string contain
func Contains(arr []string, str string) bool {
	for _, item := range arr {
		// fmt.Println("item == str", item, str, (item == str))
		if item == str {
			return true
		}
	}
	return false
}

// Exist if the path exist
func Exist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
