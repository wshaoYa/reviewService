package reviewService

import (
	"strconv"
	"unicode"
)

func decodeString(s string) string {
	var (
		st []rune
	)

	for _, c := range s {
		if c != ']' {
			st = append(st, c)
			continue
		}

		//str
		var j int
		for j = len(st) - 1; st[j] != '['; j-- {
		}
		//str := slices.Clone(st[j+1:])
		str := st[j+1:]
		st = st[:j]

		//num
		for j--; j >= 0 && unicode.IsDigit(st[j]); j-- {
		}
		num, _ := strconv.Atoi(string(st[j+1:]))
		st = st[:j+1]

		//append
		for i := 0; i < num; i++ {
			st = append(st, str...)
		}
	}

	return string(st)
}

func main() {
	decodeString("3[a2[c]]")
}
