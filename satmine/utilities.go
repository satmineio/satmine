package satmine

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
)

// GenerateRandomBitcoinBlockHash generates a random hash that simulates a Bitcoin block hash
func GenerateRandomBitcoinBlockHash() string {
	// Create a buffer of random data of sufficient size
	randomData := make([]byte, 80) // Bitcoin block header size is approximately 80 bytes

	// Use crypto/rand to generate random data to fill the buffer
	if _, err := rand.Read(randomData); err != nil {
		return ""
	}

	// Perform two rounds of SHA256 hashing on the random data
	hash := sha256.Sum256(randomData)
	hash = sha256.Sum256(hash[:])

	// Convert the hash value to a hexadecimal string
	hashStr := hex.EncodeToString(hash[:])

	return hashStr
}

// AddBigNumbers performs addition of two large numbers represented as strings.
// It converts the string inputs to big.Int, performs the addition, and returns the result as a string.
func AddBigNumbers(a, b string) (string, error) {
	// Convert string inputs to big.Int
	aBigInt, ok := new(big.Int).SetString(a, 10)
	if !ok {
		return "", fmt.Errorf("invalid input: %s", a)
	}
	bBigInt, ok := new(big.Int).SetString(b, 10)
	if !ok {
		return "", fmt.Errorf("invalid input: %s", b)
	}

	// Perform addition
	result := new(big.Int).Add(aBigInt, bBigInt)

	// Return the result as a string
	return result.String(), nil
}

// SubtractBigNumbers performs subtraction of two large numbers represented as strings.
// It converts the string inputs to big.Int, performs the subtraction, and returns the result as a string.
func SubtractBigNumbers(a, b string) (string, error) {
	// Convert string inputs to big.Int
	aBigInt, ok := new(big.Int).SetString(a, 10)
	if !ok {
		return "", fmt.Errorf("invalid input: %s", a)
	}
	bBigInt, ok := new(big.Int).SetString(b, 10)
	if !ok {
		return "", fmt.Errorf("invalid input: %s", b)
	}

	// Perform subtraction
	result := new(big.Int).Sub(aBigInt, bBigInt)

	// Return the result as a string
	return result.String(), nil
}

// convertHashToBigInt converts a hash string (hexadecimal) to a big.Int in the range of 1 to limit.
// Returns an error if the hash conversion fails.
func convertHashToBigInt(hash string, limit int64) (*big.Int, error) {
	// Check if hash starts with "0x" and remove it
	if limit == 0 {
		return nil, errors.New("convertHashToBigInt limit = 0")
	}
	if limit == 1 {
		return big.NewInt(0), nil
	}

	if len(hash) >= 2 && hash[:2] == "0x" {
		hash = hash[2:]
	}
	// Convert hash to big.Int
	hashInt, ok := new(big.Int).SetString(hash, 16)
	if !ok {
		return nil, errors.New("invalid hash format" + fmt.Sprintf(" hash= %s\n", hash))
	}

	// // Ensure the limit is a positive number
	// if limit <= 0 {
	// 	return nil, errors.New("limit must be greater than 0")
	// }

	// Modulo operation to bring the value within the range of 1 to limit
	// Add 1 to ensure the range is 1 to limit inclusive
	result := new(big.Int).Mod(hashInt, big.NewInt(limit))
	//result.Add(result, big.NewInt(1))

	return result, nil
}

// stringToPercentageBigInt converts a string representing a floating-point number
// to a big integer after multiplying it by 1000 and rounding it.
func stringToPercentageBigInt(str string) *big.Int {
	// Convert the string to a floating-point number
	fltValue, err := strconv.ParseFloat(str, 64)
	if err != nil {
		fmt.Printf("Error parsing float: %v\n", err)
		return big.NewInt(0)
	}

	// Multiply the float by 1000 and round to the nearest whole number
	multiplied := fltValue * 1000
	rounded := int64(multiplied)

	// Convert the rounded value to big.Int
	resultBigInt := big.NewInt(rounded)

	return resultBigInt
}
