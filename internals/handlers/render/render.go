package render

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// APIError wraps all informations required to investiguate a backend error
// It is mainly used to returns information to the API caller when the status is not 2xx.
type APIError struct {
	RequestID string `json:"requestID"`
	Status    int    `json:"status"`
	ErrType   string `json:"type"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
}

var (
	// ErrAPIDecodeJSONBody must be used when a JSON decoding error occurs (mainly for `json.NewDecoder(r.Body).Decode(&myObject)`)
	ErrAPIDecodeJSONBody = APIError{Status: http.StatusBadRequest, ErrType: "ParsingError", Code: 1000, Message: `Failed to parse the JSON body provided in the request`}
	// ErrAPIEncodeJSONBody must be used when a JSON encoding error occurs (mainly for `json.NewEncoder(w).Encode(myObject)`)
	ErrAPIEncodeJSONBody = APIError{Status: http.StatusBadRequest, ErrType: "ParsingError", Code: 1001, Message: `Failed to encode the JSON response`}
	// ErrAPIParsingInteger must be used when an int parsing error occurs (mainly for query parameters parsing)
	ErrAPIParsingInteger = APIError{Status: http.StatusBadRequest, ErrType: "ParsingError", Code: 1002, Message: `Failed to parse a query param of type 'integer'`}
	// ErrAPIParsingDateTime must be used when a time.Time parsing error occurs (mainly for query parameters parsing)
	ErrAPIParsingDateTime = APIError{Status: http.StatusBadRequest, ErrType: "ParsingError", Code: 1003, Message: `Failed to parse a query param of type 'datetime'. A datetime parameter must match the following parttern : "2006-01-02T15:04:05.000Z07:00" (example : "2020-06-23T15:30:01+02:00")`}
	// ErrAPIParsingDuration must be used when a time.Duration parsing error occurs (mainly for query parameters parsing)
	ErrAPIParsingDuration = APIError{Status: http.StatusBadRequest, ErrType: "ParsingError", Code: 1004, Message: `Failed to parse a query param of type 'duration'. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". Multiple duration can be combined (example: "23s", "48h", "1h30m40s"`}
	// ErrAPIParsingKeyValue must be used when a key-value type parsing error occurs (mainly for query parameters parsing, ie. placeholders)
	ErrAPIParsingKeyValue = APIError{Status: http.StatusBadRequest, ErrType: "ParsingError", Code: 1005, Message: `Failed to parse a query param of type 'keyvalue'. Valid format is "key1:value1,key2:value2,...". (example: "country:france,category:export")`}
	// ErrAPIParsingSortBy must be used when a models.SortOption type parsing error occurs (mainly for query parameters parsing)
	ErrAPIParsingSortBy = APIError{Status: http.StatusBadRequest, ErrType: "ParsingError", Code: 1006, Message: `Failed to parse a query param of type 'sort_by'. Valid format is "desc(field1),asc(field2),...". (example: "asc(date),desc(country),asc(category)`}

	// ErrAPIMissingParam must be used when a mandatory query parameter is missing
	ErrAPIMissingParam = APIError{Status: http.StatusBadRequest, ErrType: "RessourceError", Code: 2000, Message: `Query has missing parameter and cannot be processed`}
	// ErrAPIResourceInvalid must be used when the provided resource is invalid from a "business view" (ie. the JSON is valid, but it's content is not)
	ErrAPIResourceInvalid = APIError{Status: http.StatusBadRequest, ErrType: "RessourceError", Code: 2001, Message: `Provided resource definition can be parsed, but is invalid`}
	// ErrAPIResourceDuplicate must be used in case a duplicate resource has been identified
	ErrAPIResourceDuplicate = APIError{Status: http.StatusBadRequest, ErrType: "RessourceError", Code: 2002, Message: `Provided resource definition can be parsed, but is already exists`}

	// ErrAPIDBResourceNotFound must be used in case a resource is not found in the backend storage system
	ErrAPIDBResourceNotFound = APIError{Status: http.StatusNotFound, ErrType: "RessourceError", Code: 3000, Message: `Ressource not found`}
	// ErrAPIDBSelectFailed must be used when a select query returns an error from the backend storage system
	ErrAPIDBSelectFailed = APIError{Status: http.StatusInternalServerError, ErrType: "RessourceError", Code: 3001, Message: `Failed to execute the query`}
	// ErrAPIDBInsertFailed must be used when an insert query returns an error from the backend storage system
	ErrAPIDBInsertFailed = APIError{Status: http.StatusInternalServerError, ErrType: "RessourceError", Code: 3002, Message: `Failed to create the resource`}
	// ErrAPIDBUpdateFailed must be used when an update query returns an error from the backend storage system
	ErrAPIDBUpdateFailed = APIError{Status: http.StatusInternalServerError, ErrType: "RessourceError", Code: 3003, Message: `Failed to update the resource`}
	// ErrAPIDBDeleteFailed must be used when a delete query returns an error from the backend storage system
	ErrAPIDBDeleteFailed = APIError{Status: http.StatusInternalServerError, ErrType: "RessourceError", Code: 3004, Message: `Failed to delete the resource`}
	// ErrAPIDBResourceNotFoundAfterInsert must be used when a resource is not found right after an insert/update in the backend storage system
	ErrAPIDBResourceNotFoundAfterInsert = APIError{Status: http.StatusInternalServerError, ErrType: "RessourceError", Code: 3005, Message: `Ressource not found after creation`}
	// ErrAPIDBTransactionBegin must be used when a transaction beging fails
	ErrAPIDBTransactionBegin = APIError{Status: http.StatusInternalServerError, ErrType: "ProcessError", Code: 3006, Message: `Field to begin the transaction`}
	// ErrAPIDBTransactionCommit must be used when a transaction commit fails
	ErrAPIDBTransactionCommit = APIError{Status: http.StatusInternalServerError, ErrType: "ProcessError", Code: 3007, Message: `Field to commit the transaction`}

	// ErrAPIElasticSelectFailed must be used when an Elasticsearch query fails
	ErrAPIElasticSelectFailed = APIError{Status: http.StatusInternalServerError, ErrType: "RessourceError", Code: 4000, Message: `Failed to execute the query`}

	// ErrAPIProcessError must be used when an internal error occurred during the stack call
	ErrAPIProcessError = APIError{Status: http.StatusUnauthorized, ErrType: "ProcessError", Code: 5000, Message: `Internal error has occurred during the process`}

	// ErrAPISecurityMissingContext must be used in case no security context is found (missing credentials, missing jwt, etc.)
	// or the context is invalid (invalid jwt, user not found, etc.)
	// This is a specific case when the least details are added for security reason
	ErrAPISecurityMissingContext = APIError{Status: http.StatusUnauthorized, ErrType: "SecurityError", Code: 6000, Message: `Security error. Please contact an administrator`}
	// ErrAPISecurityNoRights must be used when the used is properly authenticated but doesn't have the required rights to access the resource
	ErrAPISecurityNoRights = APIError{Status: http.StatusUnauthorized, ErrType: "SecurityError", Code: 6001, Message: `Security error. Please contact an administrator`}
)

// OK returns a HTTP status 200 with an empty body
func OK(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// NotImplemented returns a HTTP status 501
func NotImplemented(w http.ResponseWriter, r *http.Request) {
	JSON(w, r, map[string]interface{}{"message": "Not Implemented"})
	w.WriteHeader(http.StatusNotImplemented)
}

// JSON try to encode an interface and returns it in a specific ResponseWriter (or returns an internal server error)
func JSON(w http.ResponseWriter, r *http.Request, data interface{}) {
	OK(w, r)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		zap.L().Error("Render JSON encode", zap.Error(err))
		Error(w, r, ErrAPIEncodeJSONBody, err)
		return
	}
}

// Error handles and return an error (JSON format) with corresponding HTTP status
// In case the API is configured with API_ENABLE_VERBOSE_ERROR = true, the detailed errors will also be sent in the JSON response
func Error(w http.ResponseWriter, r *http.Request, apiError APIError, err error) {
	apiError.RequestID = middleware.GetReqID(r.Context())

	if viper.GetBool("API_ENABLE_VERBOSE_ERROR") {
		apiError.Details = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiError.Status)

	encodeErr := json.NewEncoder(w).Encode(apiError)
	if encodeErr != nil {
		zap.L().Error("Error JSON encode", zap.Error(encodeErr))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}
