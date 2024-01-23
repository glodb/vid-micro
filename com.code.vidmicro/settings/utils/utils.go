package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const saltLength = 16

func InterfaceArrayToIntArray(floatArray []interface{}) []int {
	intArray := make([]int, len(floatArray))

	for i, v := range floatArray {
		// Convert float64 to int using rounding or other logic as needed
		intArray[i] = int(math.Round(v.(float64)))
	}

	return intArray
}

func GenerateUUID() (string, error) {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return "", err
	}
	// Set the version (4) and variant (RFC4122) bits
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf(
			"%x-%x-%x-%x-%x",
			uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]),
		nil
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

func HashPassword(password string, salt []byte) string {
	passwordBytes := []byte(password)
	combined := append(passwordBytes, salt...)
	hash := sha256.Sum256(combined)
	return base64.StdEncoding.EncodeToString(hash[:])
}

func GenerateToken() (string, error) {
	randomBytes := make([]byte, 32) // 32 bytes for SHA-256

	// Generate random bytes
	_, err := rand.Read(randomBytes)
	if err != nil {
		fmt.Println("Error generating random bytes:", err)
		return "", err
	}

	// Hash the random bytes using SHA-256
	hash := sha256.Sum256(randomBytes)

	// Convert the hash to a hexadecimal string
	hashString := hex.EncodeToString(hash[:])

	return hashString, nil
}

func GetDayStart(deviceTime int32) int32 {
	secondsToSubtract := deviceTime % (24 * 60 * 60)
	timeBracket := deviceTime - secondsToSubtract
	return timeBracket
}

func GetHourStart(deviceTime int64) int64 {
	secondsToSubtract := deviceTime % (60 * 60)
	timeBracket := deviceTime - secondsToSubtract
	return timeBracket
}

func GetMonthStart(timeUnix int32) int32 {
	timeObj := time.Unix(int64(timeUnix), 0)
	timeObj = timeObj.AddDate(0, 0, -timeObj.Day()+1)
	timeObjNew := time.Date(timeObj.Year(), timeObj.Month(), timeObj.Day(), 0, 0, 0, 0, timeObj.Location())
	return int32(timeObjNew.Unix())
}

func GetMonthEnd(timeUnix int32) int32 {
	timeObj := time.Unix(int64(timeUnix), 0)
	timeObj.AddDate(0, 1, -timeObj.Day()).Unix()
	timeObjNew := time.Date(timeObj.Year(), timeObj.Month(), timeObj.Day(), 0, 0, 0, 0, timeObj.Location())
	return int32(timeObjNew.Unix())
}

func RequestHttp(url string, method string, data string, contentType string, authorization string) (io.ReadCloser, error) {
	payload := strings.NewReader(data)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", contentType)

	if authorization != "" {
		req.Header.Add("Authorization", authorization)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

func EncodeToString(max int) string {
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		panic(err)
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}

func GetMinutesDiff(deviceTime int32, minDelay int32) int32 {
	secondsToSubtract := deviceTime % (minDelay * 60)
	timeBracket := deviceTime - secondsToSubtract
	return timeBracket
}
func GetDistance(lat1 float64, lon1 float64, lat2 float64, lon2 float64) float64 {

	R := float64(6371) // KM
	p := 0.017453292519943295
	φ1 := lat1 * p // φ, λ in radians
	φ2 := lat2 * p
	Δφ := (lat2 - lat1) * p
	Δλ := (lon2 - lon1) * p

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*
			math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c // in KMs
	//R := 6371
	//p := 0.017453292519943295 // Math.PI / 180
	//var c = Math.cos;
	//var a = 0.5 - math.Cos((lat2 - lat1) * p) / 2 + math.Cos(lat1 * p) * math.Cos(lat2 * p) * (1 - math.Cos((lon2 - lon1) * p)) / 2;
	//return 12742 * math.Asin((math.Sqrt(a))); // 2  R; R = 6371 km
}

func CopyMap(m map[string]*Queue) map[string][]interface{} {
	cp := make(map[string][]interface{})
	for k, v := range m {
		cp[k] = v.Copy()
	}

	return cp
}

func ConvertDeviceTime(date string) int64 {
	if date == "0" {
		return 0
	}
	var day int64 = 0
	var month int64 = 0
	var year int64 = 0
	if len(date) >= 6 {
		day, _ = strconv.ParseInt(date[4:6], 10, 32)
		month, _ = strconv.ParseInt(date[2:4], 10, 32)
		year, _ = strconv.ParseInt(("20" + date[0:2]), 10, 32)
	}
	var hour int64 = 0
	var min int64 = 0
	var sec int64 = 0
	if len(date) > 10 {
		hour, _ = strconv.ParseInt(date[6:8], 10, 32)
		min, _ = strconv.ParseInt(date[8:10], 10, 32)
		sec, _ = strconv.ParseInt(date[10:], 10, 32)
	}
	timestamp := time.Date(int(year), time.Month(month), int(day), int(hour), int(min), int(sec), 0, time.Now().Location())
	return timestamp.Unix()
}

func ReadUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}
func AppendIfMissing(slice []string, i string) []string {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}
func IsMobileDevice(r *http.Request) bool {
	userAgetnt := r.UserAgent()
	deviceName := []byte("Android|iPhone|iPad|iPod|IOS")
	ok, _ := regexp.Match(userAgetnt, deviceName)

	return ok
}
