package helper

import (
	"crypto/rand"
	"strconv"
	"strings"
)

const numbers = "1234567890"

func RandomNumbers(length int) (int, error) {
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return 0, err
	}
	numLength := len(numbers)
	for i := 0; i < length; i++ {
		buffer[i] = numbers[int(buffer[i])%numLength]
	}
	return strconv.Atoi(string(buffer))
}

// FormatPhoneToE164 formats a phone number to E.164 format with country code
// If the number already starts with +, it returns it as is (assuming already formatted)
// Otherwise, it adds the country code (+234 for Nigeria by default)
func FormatPhoneToE164(phone string) string {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return phone
	}

	// If already in E.164 format (starts with +), return as is
	if strings.HasPrefix(phone, "+") {
		return phone
	}

	// Remove any spaces, dashes, or other formatting
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")

	// Remove leading zero if present (common in Nigerian numbers)
	phone = strings.TrimPrefix(phone, "0")

	// If starts with country code without +, add +
	if strings.HasPrefix(phone, "234") {
		return "+" + phone
	}

	// Default: assume Nigerian number and add +234
	return "+234" + phone
}
