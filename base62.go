package main

import (
	"bytes"
	"math"
)

// characters used for conversion
const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// converts number to base62
func Encode(number int64) string {
	if number == 0 {
		return string(alphabet[0])
	}

	chars := make([]byte, 0)

	length := len(alphabet)

	for number > 0 {
		result := number / int64(length)
		remainder := number % int64(length)
		chars = append(chars, alphabet[remainder])
		number = result
	}

	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}

	return string(chars)
}

// converts base62 token to int
func Decode(token string) int64 {
	var number int64 = 0
	idx := 0.0
	chars := []byte(alphabet)

	charsLength := float64(len(chars))
	tokenLength := float64(len(token))

	for _, c := range []byte(token) {
		power := tokenLength - (idx + 1)
		index := bytes.IndexByte(chars, c)
		number += int64(index) * int64(math.Pow(charsLength, power))
		idx++
	}

	return number
}
