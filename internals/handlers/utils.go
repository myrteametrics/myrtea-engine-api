package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/utils"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"

	"net/http"
	"net/url"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/users"
	"go.uber.org/zap"
)

const (
	ExpectedStateErr            = "Error to generate a random State for OIDC Authentification"
	InvalidStateErr             = "OIDC authentication invalid state"
	TokenExchangeErr            = "OIDC authentication Failed to exchange token"
	NoIDTokenErr                = "OIDC authentication No ID token found"
	IDTokenVerifyErr            = "OIDC authentication Failed to verify ID token"
	TokenName                   = "token"
	AllowedCookiePath           = "/"
	EnableParsingQueryParamName = "parsinggvalenabled"
)

// QueryParamToOptionalInt parse a string from a string
func QueryParamToOptionalInt(r *http.Request, name string, orDefault int) (int, error) {
	param := r.URL.Query().Get(name)
	if param != "" {
		return strconv.Atoi(param)
	}
	return orDefault, nil
}

// QueryParamToOptionalInt64 parse an int64 from a string
func QueryParamToOptionalInt64(r *http.Request, name string, orDefault int64) (int64, error) {
	param := r.URL.Query().Get(name)
	if param != "" {
		return strconv.ParseInt(param, 10, 64)
	}
	return orDefault, nil
}

// QueryParamToOptionalInt64Array parse multiple int64 entries separated by a separator from a string
func QueryParamToOptionalInt64Array(r *http.Request, name, separator string, allowDuplicates bool, orDefault []int64) ([]int64, error) {
	param := r.URL.Query().Get(name)
	if param == "" {
		return orDefault, nil
	}
	split := strings.Split(param, separator)
	result := make([]int64, len(split))

	for i := 0; i < len(split); i++ {
		val, err := strconv.ParseInt(split[i], 10, 64)
		if err != nil {
			return nil, err
		}
		result[i] = val
	}

	if !allowDuplicates {
		return utils.RemoveDuplicates(result), nil
	}

	return result, nil
}

func QueryParamToOptionalStringArray(r *http.Request, name string, separator string, orDefault []string) []string {
	param := r.URL.Query().Get(name)
	if param != "" {
		return strings.Split(param, separator)
	}
	return orDefault
}

func QueryParamToOptionalTime(r *http.Request, name string, orDefault time.Time) (time.Time, error) {
	param := r.URL.Query().Get(name)
	if param != "" {
		return time.Parse("2006-01-02T15:04:05.000Z07:00", param)
	}
	return orDefault, nil
}

func QueryParamToTime(r *http.Request, name string) (time.Time, error) {
	param := r.URL.Query().Get(name)
	if param != "" {
		return time.Parse("2006-01-02T15:04:05.000Z07:00", param)
	}
	return time.Time{}, fmt.Errorf("missing query parameter %s", name)
}

func QueryParamToOptionalDuration(r *http.Request, name string, orDefault time.Duration) (time.Duration, error) {
	param := r.URL.Query().Get(name)
	if param != "" {
		return time.ParseDuration(param)
	}
	return orDefault, nil
}

func QueryParamToOptionalBool(r *http.Request, name string, orDefault bool) (bool, error) {
	param := r.URL.Query().Get(name)
	if param != "" {
		return strconv.ParseBool(param)
	}
	return orDefault, nil
}

// QueryParamToOptionalKeyValues parse multiple key:value entries separated by commas from a string
func QueryParamToOptionalKeyValues(r *http.Request, name string, orDefault map[string]string) (map[string]string, error) {
	param := r.URL.Query().Get(name)
	if param != "" {
		keyValues := make(map[string]string)
		rawKeyValues := strings.Split(param, ",")
		for _, rawKeyValue := range rawKeyValues {
			keyValue := strings.Split(rawKeyValue, ":")
			if len(keyValue) != 2 {
				return nil, fmt.Errorf("invalid placeholder key-value : %s", keyValue)
			}
			keyValues[keyValue[0]] = keyValue[1]
		}
		return keyValues, nil
	}
	return orDefault, nil
}

// ParseTime try to parse a supposed time string as a time.Time or returns time.Now()
func ParseTime(tStr string) (time.Time, error) {
	t, err := time.Parse("2006-01-02T15:04:05.000Z07:00", tStr)
	if err != nil {
		return time.Now().UTC(), err
	}
	return t, nil
}

// ParseInt try to parse a string as an int or returns 0
func ParseInt(tStr string) (int, error) {
	if tStr != "" {
		i, err := strconv.Atoi(tStr)
		if err != nil {
			return 0, err
		}
		return i, nil
	}
	return 0, nil
}

// ParseDuration try to parse a string as an int or returns 0
func ParseDuration(dStr string) (time.Duration, error) {
	if dStr != "" {
		d, err := time.ParseDuration(dStr)
		if err != nil {
			return 0, err
		}
		return d, nil
	}
	return 0, nil
}

// sortByRegex is a regex matching expression similar to <order>(<field>)
// where <order> must be 'asc' or 'desc' and <field> cannot contains parenthesis
var sortByRegex = regexp.MustCompile(`^(asc|desc)\(([A-Za-z0-9_]+?)\)$`)

// ParseSortBy parse multiple <order>(<field>) entries separated by commas from a string
func ParseSortBy(rawSortByStr string, allowedFields []string) ([]models.SortOption, error) {
	sortOptions := make([]models.SortOption, 0)
	for _, sortByStr := range strings.Split(rawSortByStr, ",") {
		sortByStr = strings.TrimSpace(sortByStr)
		if sortByStr == "" {
			continue
		}

		parsing := sortByRegex.FindStringSubmatch(sortByStr)
		if len(parsing) < 3 {
			return nil, fmt.Errorf("invalid sortby clause '%s'", sortByStr)
		}

		order := models.ToSortOptionOrder(parsing[1])
		if order == 0 {
			return nil, fmt.Errorf("invalid sortby order found '%s'", parsing[1])
		}

		field := parsing[2]
		found := false
		for _, allowedField := range allowedFields {
			if field == allowedField {
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("invalid sortby field found '%s'", field)
		}

		sortOptions = append(sortOptions, models.SortOption{
			Field: field,
			Order: order,
		})
	}
	return sortOptions, nil
}

// GetUserFromContext extract the logged user from the request context
func GetUserFromContext(r *http.Request) (users.UserWithPermissions, bool) {
	c := r.Context()
	_user := c.Value(models.ContextKeyUser)
	if _user == nil {
		zap.L().Warn("No context user provided")
		return users.UserWithPermissions{}, false
	}
	user := _user.(users.UserWithPermissions)
	return user, true
}

// handleError is a helper function that logs the error and sends a response.
func handleError(w http.ResponseWriter, r *http.Request, message string, err error, apiError render.APIError) {
	zap.L().Error(message, zap.Error(err))
	render.Error(w, r, apiError, err)
}

// generateRandomState Generate a State used by OIDC authentication
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// generateEncryptedState Generate a State used by OIDC authentication
func generateEncryptedState(key []byte) (string, error) {
	// Generate random state
	plainState, err := generateRandomState()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plainState))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plainState))

	// Encode to base64
	b64State := base64.StdEncoding.EncodeToString(ciphertext)
	return b64State, nil
}

// verifyEncryptedState Verify the State used by OIDC authentication
func verifyEncryptedState(state string, key []byte) (string, error) {
	// Decode from base64
	decodedState, err := base64.StdEncoding.DecodeString(state)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(decodedState) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	iv := decodedState[:aes.BlockSize]
	decodedState = decodedState[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(decodedState, decodedState)

	return string(decodedState), nil
}

// findCombineFacts returns the combine facts
func findCombineFacts(combineFactIds []int64) (combineFacts []engine.Fact) {
	for _, factId := range utils.RemoveDuplicates(combineFactIds) {
		combineFact, found, err := fact.R().Get(factId)
		if err != nil {
			continue
		}
		if !found {
			continue
		}
		combineFacts = append(combineFacts, combineFact)
	}
	return combineFacts
}

func gvalParsingEnabled(params url.Values) bool {
	val := params.Get(EnableParsingQueryParamName)
	if val == "" {
		return false
	}
	parsedVal, err := strconv.ParseBool(val)
	if err != nil {
		return false
	}
	return parsedVal
}

// ParsefactParameters takes a string of encoded parameters and returns a map of these decoded parameters.
func ParseFactParameters(factParameters string) (map[string]string, error) {

	if factParameters == "" {
		return make(map[string]string), nil
	}

	decodedValue, err := url.QueryUnescape(factParameters)

	if err != nil {
		return nil, fmt.Errorf("cannot decode: %v", err)
	}

	paramsMap := make(map[string]string)

	//Separation of key pairs = value
	pairs := strings.Split(decodedValue, "&")
	for _, pair := range pairs {
		// Separation of key and value
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid pair: %v", pair)
		}

		paramsMap[kv[0]] = kv[1]
	}

	return paramsMap, nil
}
