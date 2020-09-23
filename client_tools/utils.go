package main

import (
	"reflect"
)

func Find(collection []string, value string) int {
	for i := 0; i < len(collection); i++ {
		if collection[i] == value {
			return i
		}
	}
	return -1
}
func Contains(collection []string, value string) bool {
	return Find(collection, value) != -1
}

func FindIf(slice interface{}, pred func(i int) bool) int {
	in := reflect.ValueOf(slice)
	n := in.Len()
	for i := 0; i < n; i++ {
		if pred(i) {
			return i
		}
	}
	return -1
}
