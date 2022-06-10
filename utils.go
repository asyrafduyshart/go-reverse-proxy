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

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func FindAndDelete(s []string, item string) []string {
	index := 0
	for _, i := range s {
		if i != item {
			s[index] = i
			index++
		}
	}
	return s[:index]
}
