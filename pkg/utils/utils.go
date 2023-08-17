package utils

func RemoveAtIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}
