package utils

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"
)

func IsSlice(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Slice
}

func IsStringInSlice(slice []string, v string) bool {
	for _, s := range slice {
		if s == v {
			return true
		}
	}
	return false
}

func InterfaceToSlice(t []interface{}) []string {
	s := make([]string, len(t))
	for i, v := range t {
		s[i] = fmt.Sprint(v)
	}
	return s
}

func GetUserInput(p string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(p)
	res, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	res = strings.TrimSuffix(res, "\n")
	res = strings.TrimSuffix(res, "\r")

	return res, nil
}
