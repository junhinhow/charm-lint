package rangeindex

// previousIsSpace compara o offset de byte com zero e com o byte anterior.
func previousIsSpace(s string) int {
	count := 0
	for i := range s {
		if i > 0 { // want `offset de byte`
			if s[i-1] == ' ' { // want `offset de byte`
				count++
			}
		}
	}
	return count
}

// byteAfter soma um ao offset para ler o próximo byte.
func byteAfter(s string) byte {
	for i := range s {
		return s[i+1] // want `offset de byte`
	}
	return 0
}

// lastIndex usa o offset em uma comparação de igualdade.
func lastIndex(s string) bool {
	for i := range s {
		return i == len(s)-1 // want `offset de byte`
	}
	return false
}
