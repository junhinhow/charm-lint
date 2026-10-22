package rangeindex

// runeIndex itera sobre runes: i aqui já é o índice do rune.
func runeIndex(s string) int {
	total := 0
	for i, r := range s {
		total += i + int(r)
	}
	return total
}

// byteIndex itera sobre []byte: a aritmética é de bytes de propósito.
func byteIndex(b []byte) int {
	total := 0
	for i := range b {
		total += i + 1
	}
	return total
}

// classicLoop não usa range.
func classicLoop(s string) int {
	total := 0
	for i := 0; i < len(s); i++ {
		total += i
	}
	return total
}
