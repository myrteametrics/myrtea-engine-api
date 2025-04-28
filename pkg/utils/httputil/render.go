package httputil

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// APIError wraps all information required to investigate a backend error
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
	// ErrAPIParsingUUID must be used when a UUID parsing error occurs (mainly for query parameters parsing)
	ErrAPIParsingUUID = APIError{Status: http.StatusBadRequest, ErrType: "ParsingError", Code: 1007, Message: `Failed to parse a query param of type 'uuid'. Valid format is "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx". (example: "123e4567-e89b-12d3-a456-426614174000")`}

	// ErrAPIMissingParam must be used when a mandatory query parameter is missing
	ErrAPIMissingParam = APIError{Status: http.StatusBadRequest, ErrType: "ResourceError", Code: 2000, Message: `Query has missing parameter and cannot be processed`}
	// ErrAPIUnexpectedParamValue must be used when a parameter value is not expected
	ErrAPIUnexpectedParamValue = APIError{Status: http.StatusBadRequest, ErrType: "ResourceError", Code: 2000, Message: `Query has unexpected parameter value and cannot be processed`}
	// ErrAPIResourceInvalid must be used when the provided resource is invalid from a "business view" (ie. the JSON is valid, but it's content is not)
	ErrAPIResourceInvalid = APIError{Status: http.StatusBadRequest, ErrType: "ResourceError", Code: 2001, Message: `Provided resource definition can be parsed, but is invalid`}
	// ErrAPIResourceDuplicate must be used in case a duplicate resource has been identified
	ErrAPIResourceDuplicate = APIError{Status: http.StatusBadRequest, ErrType: "ResourceError", Code: 2002, Message: `Provided resource definition can be parsed, but is already exists`}

	// ErrAPIQueueFull must be used in case an internal processing queue is full
	ErrAPIQueueFull = APIError{Status: http.StatusServiceUnavailable, ErrType: "ResourceError", Code: 2003, Message: `The queue is full, please retry later`}

	// ErrAPIElasticExportDisabled must be used in case an custom export is asked and export is disabled in config
	ErrAPIElasticExportDisabled = APIError{Status: http.StatusForbidden, ErrType: "ResourceError", Code: 2004, Message: `Export is disabled on the selected elastic config`}

	// ErrAPITooManyRequests must be used in case the client has sent too many requests in a given amount of time
	ErrAPITooManyRequests = APIError{Status: http.StatusTooManyRequests, ErrType: "ResourceError", Code: 2004, Message: `Too many requests, please try again later`}

	// ErrAPIDBResourceNotFound must be used in case a resource is not found in the backend storage system
	ErrAPIDBResourceNotFound = APIError{Status: http.StatusNotFound, ErrType: "ResourceError", Code: 3000, Message: `Ressource not found`}
	// ErrAPIDBSelectFailed must be used when a select query returns an error from the backend storage system
	ErrAPIDBSelectFailed = APIError{Status: http.StatusInternalServerError, ErrType: "ResourceError", Code: 3001, Message: `Failed to execute the query`}
	// ErrAPIDBInsertFailed must be used when an insert query returns an error from the backend storage system
	ErrAPIDBInsertFailed = APIError{Status: http.StatusInternalServerError, ErrType: "ResourceError", Code: 3002, Message: `Failed to create the resource`}
	// ErrAPIDBUpdateFailed must be used when an update query returns an error from the backend storage system
	ErrAPIDBUpdateFailed = APIError{Status: http.StatusInternalServerError, ErrType: "ResourceError", Code: 3003, Message: `Failed to update the resource`}
	// ErrAPIDBDeleteFailed must be used when a delete query returns an error from the backend storage system
	ErrAPIDBDeleteFailed = APIError{Status: http.StatusInternalServerError, ErrType: "ResourceError", Code: 3004, Message: `Failed to delete the resource`}
	// ErrAPIDBResourceNotFoundAfterInsert must be used when a resource is not found right after an insert/update in the backend storage system
	ErrAPIDBResourceNotFoundAfterInsert = APIError{Status: http.StatusInternalServerError, ErrType: "ResourceError", Code: 3005, Message: `Ressource not found after creation`}
	// ErrAPIDBTransactionBegin must be used when a transaction beging fails
	ErrAPIDBTransactionBegin = APIError{Status: http.StatusInternalServerError, ErrType: "ProcessError", Code: 3006, Message: `Field to begin the transaction`}
	// ErrAPIDBTransactionCommit must be used when a transaction commit fails
	ErrAPIDBTransactionCommit = APIError{Status: http.StatusInternalServerError, ErrType: "ProcessError", Code: 3007, Message: `Field to commit the transaction`}

	// ErrAPIElasticSelectFailed must be used when an Elasticsearch query fails
	ErrAPIElasticSelectFailed = APIError{Status: http.StatusInternalServerError, ErrType: "ResourceError", Code: 4000, Message: `Failed to execute the query`}

	// ErrAPIProcessError must be used when an internal error occurred during the stack call
	ErrAPIProcessError = APIError{Status: http.StatusInternalServerError, ErrType: "ProcessError", Code: 5000, Message: `Internal error has occurred during the process`}

	// ErrAPISecurityMissingContext must be used in case no security context is found (missing credentials, missing jwt, etc.)
	// or the context is invalid (invalid jwt, user not found, etc.)
	// This is a specific case when the least details are added for security reason
	ErrAPISecurityMissingContext = APIError{Status: http.StatusUnauthorized, ErrType: "SecurityError", Code: 6000, Message: `Security error. Please contact an administrator`}
	// ErrAPISecurityNoPermissions must be used when the user is properly authenticated but doesn't have the required rights to access the resource
	ErrAPISecurityNoPermissions = APIError{Status: http.StatusForbidden, ErrType: "SecurityError", Code: 6001, Message: `Security error. Please contact an administrator`}

	// ErrAPIPartialSuccess must be used when some requests succeeded, and others failed.
	// This error is used to indicate that the processing of a batch of requests has
	// encountered issues, but not all of them have failed. The error should provide
	// information about the requests that succeeded and those that failed.
	ErrAPIPartialSuccess = APIError{Status: http.StatusPartialContent, ErrType: "MixedResult", Code: 7000, Message: `Some requests have succeeded, while others have failed.`}
	// ErrAPIGenerateRandomStateFailed should be used when the generation of a random state fails.
	ErrAPIGenerateRandomStateFailed = APIError{Status: http.StatusInternalServerError, ErrType: "ProcessError", Code: 7000, Message: `Failed to generate a random state`}
	// ErrAPIInvalidOIDCState should be used when the 'state' in the OIDC callback does not match the expected state.
	ErrAPIInvalidOIDCState = APIError{Status: http.StatusBadRequest, ErrType: "SecurityError", Code: 7002, Message: `Invalid OIDC state. The state parameter in the callback does not match the expected state`}
	// ErrAPIExchangeOIDCTokenFailed should be used when the exchange of the OIDC token fails.
	ErrAPIExchangeOIDCTokenFailed = APIError{Status: http.StatusInternalServerError, ErrType: "ProcessError", Code: 7003, Message: `Failed to exchange OIDC token`}
	// ErrAPINoIDOIDCToken should be used when no ID Token is found.
	ErrAPINoIDOIDCToken = APIError{Status: http.StatusInternalServerError, ErrType: "ProcessError", Code: 7004, Message: `No ID token found`}
	// ErrAPIVerifyIDOIDCTokenFailed should be used when the verification of the ID Token fails.
	ErrAPIVerifyIDOIDCTokenFailed = APIError{Status: http.StatusInternalServerError, ErrType: "ProcessError", Code: 7005, Message: `Failed to verify ID token`}
	// ErrAPIMissingAuthCookie should be used when an authentication cookie is missing.
	ErrAPIMissingAuthCookie = APIError{Status: http.StatusUnauthorized, ErrType: "ProcessError", Code: 7006, Message: `Missing auth cookie`}
	// ErrAPIInvalidAuthToken should be used when an authentication token is invalid.
	ErrAPIInvalidAuthToken = APIError{Status: http.StatusUnauthorized, ErrType: "ProcessError", Code: 7007, Message: `Invalid auth token`}
	// ErrAPIExpiredAuthToken should be used when an authentication token has expired.
	ErrAPIExpiredAuthToken = APIError{Status: http.StatusUnauthorized, ErrType: "ProcessError", Code: 7008, Message: `Expired auth token`}
	// ErrAPIMissingIDTokenFromContext should be used when an ID Token is missing from the context.
	ErrAPIMissingIDTokenFromContext = APIError{Status: http.StatusUnauthorized, ErrType: "ProcessError", Code: 7009, Message: `Missing idToken from context`}
	// ErrAPIFailedToGetUserClaims should be used when retrieval of user claims fails.
	ErrAPIFailedToGetUserClaims = APIError{Status: http.StatusInternalServerError, ErrType: "ProcessError", Code: 7010, Message: `Failed to get user claims`}
	// ErrAPIExportSeparatorConflict must be used when exporting a CSV file where the column separator and the list separator within a column are the same (e.g., both are commas). This configuration is not allowed as it causes ambiguity in parsing the CSV, where it becomes unclear whether a comma separates columns or list items within a column.
	ErrAPIExportSeparatorConflict = APIError{Status: http.StatusInternalServerError, ErrType: "ProcessError", Code: 7011, Message: `CSV file separator column and list separator in a column cannot be the same`}
)

// displayedErrorDetails prints details to the client
var displayedErrorDetails = []APIError{
	ErrAPIResourceInvalid,
}

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
// In case the API is configured with HTTP_SERVER_API_ENABLE_VERBOSE_ERROR = true, the detailed errors will also be sent in the JSON response
func Error(w http.ResponseWriter, r *http.Request, apiError APIError, err error) {
	apiError.RequestID = middleware.GetReqID(r.Context())

	if err != nil {
		if viper.GetBool("HTTP_SERVER_API_ENABLE_VERBOSE_ERROR") {
			apiError.Details = err.Error()
		} else {
			for _, detail := range displayedErrorDetails {
				if detail.Code == apiError.Code {
					apiError.Details = err.Error()
					break
				}
			}
		}
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

// File handles file responses which allows the download of a filename containing data
func File(w http.ResponseWriter, filename string, data []byte) {
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	_, err := w.Write(data)
	if err != nil {
		zap.L().Error("Error during file write to http response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// StreamFile handle files streamed response with allows the download of a file in chunks
func StreamFile(filePath, fileName string, w http.ResponseWriter, r *http.Request) {
	file, err := os.Open(filePath)
	if err != nil {
		Error(w, r, ErrAPIDBResourceNotFound, fmt.Errorf("error opening file: %s", err))
		return
	}
	defer file.Close()

	// Set all necessary headers
	w.Header().Set("Connection", "Keep-Alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(fileName))
	w.Header().Set("Content-Type", "application/octet-stream")

	const bufferSize = 4096
	buffer := make([]byte, bufferSize)

	for {
		// Read a chunk of the file
		bytesRead, err := file.Read(buffer)
		if err == io.EOF {
			break
		} else if err != nil {
			Error(w, r, ErrAPIProcessError, fmt.Errorf("error reading file: %s", err))
			return
		}

		// Write the chunk to the response writer
		_, err = w.Write(buffer[:bytesRead])
		if err != nil {
			// If writing to the response writer fails, log the error and stop streaming
			Error(w, r, ErrAPIProcessError, fmt.Errorf("error writing to response writer: %s", err))
			break
		}

		w.(http.Flusher).Flush()
	}
}
