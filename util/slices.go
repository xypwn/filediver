package util

// Deletes all consecutive duplicate elements, like Unix uniq.
func Uniq[S ~[]E, E comparable](s S) S {
	i := 0
	for ; i < len(s)-1; i++ {
		if s[i] == s[i+1] {
			break
		}
	}
	if i == len(s)-1 {
		return s // no duplicate found
	}

	// i: write idx
	// j: read idx
	j := i
	for ; j < len(s)-1; j++ {
		if s[j] != s[j+1] {
			s[i] = s[j]
			i++
		}
	}
	if j < len(s) {
		s[i] = s[j]
		i++
		j++
	}
	clear(s[i:])
	return s[:i]
}

// Like [Uniq], but takes a function to check equality.
func UniqFunc[S ~[]E, E any](s S, equal func(E, E) bool) S {
	i := 0
	for ; i < len(s)-1; i++ {
		if equal(s[i], s[i+1]) {
			break
		}
	}
	if i == len(s)-1 {
		return s // no duplicate found
	}

	// i: write idx
	// j: read idx
	j := i
	for ; j < len(s)-1; j++ {
		if !equal(s[j], s[j+1]) {
			s[i] = s[j]
			i++
		}
	}
	if j < len(s) {
		s[i] = s[j]
		i++
		j++
	}
	clear(s[i:])
	return s[:i]
}
