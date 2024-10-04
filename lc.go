package reviewService

func isValid(s string) bool {
	var (
		n  = len(s)
		mp = map[rune]rune{')': '(', '}': '{', ']': '['}
		st = make([]rune, 0, n)
	)

	if n%2 == 1 {
		return false
	}

	for _, c := range s {
		if v := mp[c]; v == 0 {
			st = append(st, c)
		} else {
			if len(st) == 0 || st[len(st)-1] != v {
				return false
			}
			st = st[:len(st)-1]
		}
	}
	return len(st) == 0
}
