package validation

import "unicode"

func OnlyDigits(value string) string {
	digits := make([]rune, 0, len(value))
	for _, char := range value {
		if unicode.IsDigit(char) {
			digits = append(digits, char)
		}
	}
	return string(digits)
}

func IsCPF(value string) bool {
	cpf := OnlyDigits(value)
	if len(cpf) != 11 {
		return false
	}
	allEqual := true
	for i := 1; i < len(cpf); i++ {
		if cpf[i] != cpf[0] {
			allEqual = false
			break
		}
	}
	if allEqual {
		return false
	}
	return validateDigit(cpf, 9) && validateDigit(cpf, 10)
}

func validateDigit(cpf string, position int) bool {
	sum := 0
	for i := 0; i < position; i++ {
		sum += int(cpf[i]-'0') * (position + 1 - i)
	}
	digit := (sum * 10) % 11
	if digit == 10 {
		digit = 0
	}
	return digit == int(cpf[position]-'0')
}
