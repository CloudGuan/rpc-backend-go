package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func PathExits(path string) bool {
	finfo, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}

	return finfo.IsDir()
}

func FileExits(file string) bool {
	finfo, err := os.Stat(file)
	if err != nil {
		return os.IsExist(err)
	}

	return !finfo.IsDir()
}

func DealInputPath(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		fmt.Printf("Convert %s to Abs path err %v", path, err)
		return "", err
	}

	path = strings.Replace(path, "\\", "/", -1)
	if path[len(path)-1] == '/' {
		path = path[:]
	}
	return path, nil
}

func StartWithUppercase(word string) bool {
	if len(word) == 0 {
		return false
	}
	head := word[0]
	if head >= 'A' && head <= 'Z' {
		return true
	}
	return false
}

func DealPbStructField(field string) string {
	idx := strings.Index(field, "_")
	if idx == -1 {
		return strings.ToUpper(field[:1]) + field[1:]
	}

	fls := strings.Split(field, "_")
	res := ""
	for _, fl := range fls {
		if fl == "" {
			if res == "" {
				res = "X"
			} else {
				res = res + "_"
			}
		} else if (fl[0] >= 'a' && fl[0] <= 'z') || (fl[0] >= 'A' && fl[0] <= 'Z') {
			res = res + strings.ToUpper(fl[:1]) + fl[1:]
		} else {
			res = res + "_" + fl
		}
	}
	return res
}
