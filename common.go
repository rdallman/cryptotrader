package main

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"
)

const (
	HASH_SHA1 = iota
	HASH_SHA256
	HASH_SHA512
	HASH_SHA512_384
	SATOSHIS_PER_BTC = 100000000
	SATOSHIS_PER_LTC = 100000000
)

func GetMD5(input []byte) []byte {
	hash := md5.New()
	hash.Write(input)
	return hash.Sum(nil)
}

func GetSHA512(input []byte) []byte {
	sha := sha512.New()
	sha.Write(input)
	return sha.Sum(nil)
}

func GetSHA256(input []byte) []byte {
	sha := sha256.New()
	sha.Write(input)
	return sha.Sum(nil)
}

func GetHMAC(hashType int, input, key []byte) []byte {
	var hash func() hash.Hash

	switch hashType {
	case HASH_SHA1:
		{
			hash = sha1.New
		}
	case HASH_SHA256:
		{
			hash = sha256.New
		}
	case HASH_SHA512:
		{
			hash = sha512.New
		}
	case HASH_SHA512_384:
		{
			hash = sha512.New384
		}
	}

	hmac := hmac.New(hash, []byte(key))
	hmac.Write(input)
	return hmac.Sum(nil)
}

func HexEncodeToString(input []byte) string {
	return hex.EncodeToString(input)
}

func Base64Decode(input string) ([]byte, error) {
	result, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func Base64Encode(input []byte) string {
	return base64.StdEncoding.EncodeToString(input)
}

func StringSliceDifference(slice1 []string, slice2 []string) []string {
	var diff []string
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			if !found {
				diff = append(diff, s1)
			}
		}
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}
	return diff
}

func StringContains(input, substring string) bool {
	return strings.Contains(input, substring)
}

func JoinStrings(input []string, seperator string) string {
	return strings.Join(input, seperator)
}

func SplitStrings(input, seperator string) []string {
	return strings.Split(input, seperator)
}

func TrimString(input, cutset string) string {
	return strings.Trim(input, cutset)
}

func StringToUpper(input string) string {
	return strings.ToUpper(input)
}

func StringToLower(input string) string {
	return strings.ToLower(input)
}

func RoundFloat(x float64, prec int) float64 {
	var rounder float64
	pow := math.Pow(10, float64(prec))
	intermed := x * pow
	_, frac := math.Modf(intermed)
	intermed += .5
	x = .5
	if frac < 0.0 {
		x = -.5
		intermed -= 1
	}
	if frac >= x {
		rounder = math.Ceil(intermed)
	} else {
		rounder = math.Floor(intermed)
	}

	return rounder / pow
}

func IsEnabled(isEnabled bool) string {
	if isEnabled {
		return "Enabled"
	} else {
		return "Disabled"
	}
}

func CalculateAmountWithFee(amount, fee float64) float64 {
	return amount + CalculateFee(amount, fee)
}

func CalculateFee(amount, fee float64) float64 {
	return amount * (fee / 100)
}

func CalculatePercentageDifference(amount, secondAmount float64) float64 {
	return (secondAmount - amount) / amount * 100
}

func CalculateNetProfit(amount, priceThen, priceNow, costs float64) float64 {
	return (priceNow * amount) - (priceThen * amount) - costs
}

func SendHTTPRequest(method, path string, headers map[string]string, body io.Reader) (string, error) {
	result := strings.ToUpper(method)

	if result != "POST" && result != "GET" && result != "DELETE" {
		return "", errors.New("Invalid HTTP method specified.")
	}

	req, err := http.NewRequest(method, path, body)

	if err != nil {
		return "", err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)

	if err != nil {
		return "", err
	}

	contents, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return "", err
	}

	return string(contents), nil
}

func SendHTTPGetRequest(url string, jsonDecode bool, result interface{}) (err error) {
	res, err := http.Get(url)

	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		log.Printf("HTTP status code: %d\n", res.StatusCode)
		return errors.New("Status code was not 200.")
	}

	contents, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if jsonDecode {
		err := JSONDecode(contents, &result)

		if err != nil {
			return err
		}
	} else {
		result = &contents
	}

	return nil
}

func JSONEncode(v interface{}) ([]byte, error) {
	json, err := json.Marshal(&v)

	if err != nil {
		return nil, err
	}

	return json, nil
}

func JSONDecode(data []byte, to interface{}) error {
	err := json.Unmarshal(data, &to)

	if err != nil {
		return fmt.Errorf("error unmarshaling json: %v text: %s", err, data)
	}

	return nil
}

func EncodeURLValues(url string, values url.Values) string {
	path := url
	if len(values) > 0 {
		path += "?" + values.Encode()
	}
	return path
}
