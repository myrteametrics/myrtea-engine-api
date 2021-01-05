package router

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
)

var _ samlsp.RequestTracker = CustomCookieRequestTracker{}

// CustomRequestTracker returns a new RequestTracker for the provided options,
// a CustomCookieRequestTracker which uses cookies to track pending requests.
func CustomRequestTracker(opts samlsp.Options, serviceProvider *saml.ServiceProvider) CustomCookieRequestTracker {
	return CustomCookieRequestTracker{
		ServiceProvider: serviceProvider,
		NamePrefix:      "saml_",
		Codec:           samlsp.DefaultTrackedRequestCodec(opts),
		MaxAge:          saml.MaxIssueDelay,
	}
}

// CustomCookieRequestTracker tracks requests by setting a uniquely named
// cookie for each request.
type CustomCookieRequestTracker struct {
	ServiceProvider *saml.ServiceProvider
	NamePrefix      string
	Codec           samlsp.TrackedRequestCodec
	MaxAge          time.Duration
}

// TrackRequest starts tracking the SAML request with the given ID. It returns an
// `index` that should be used as the RelayState in the SAMl request flow.
func (t CustomCookieRequestTracker) TrackRequest(w http.ResponseWriter, r *http.Request, samlRequestID string) (string, error) {
	trackedRequest := samlsp.TrackedRequest{
		Index:         base64.RawURLEncoding.EncodeToString(randomBytes(42)),
		SAMLRequestID: samlRequestID,
		// URI:           r.URL.String(),
		URI: r.Referer(),
	}
	signedTrackedRequest, err := t.Codec.Encode(trackedRequest)
	if err != nil {
		return "", err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     t.NamePrefix + trackedRequest.Index,
		Value:    signedTrackedRequest,
		MaxAge:   int(t.MaxAge.Seconds()),
		HttpOnly: true,
		Secure:   t.ServiceProvider.AcsURL.Scheme == "https",
		Path:     t.ServiceProvider.AcsURL.Path,
	})

	return trackedRequest.Index, nil
}

// StopTrackingRequest stops tracking the SAML request given by index, which is a string
// previously returned from TrackRequest
func (t CustomCookieRequestTracker) StopTrackingRequest(w http.ResponseWriter, r *http.Request, index string) error {
	cookie, err := r.Cookie(t.NamePrefix + index)
	if err != nil {
		return err
	}
	cookie.Value = ""
	cookie.Expires = time.Unix(1, 0) // past time as close to epoch as possible, but not zero time.Time{}
	http.SetCookie(w, cookie)
	return nil
}

// GetTrackedRequests returns all the pending tracked requests
func (t CustomCookieRequestTracker) GetTrackedRequests(r *http.Request) []samlsp.TrackedRequest {
	rv := []samlsp.TrackedRequest{}
	for _, cookie := range r.Cookies() {
		if !strings.HasPrefix(cookie.Name, t.NamePrefix) {
			continue
		}

		trackedRequest, err := t.Codec.Decode(cookie.Value)
		if err != nil {
			continue
		}
		index := strings.TrimPrefix(cookie.Name, t.NamePrefix)
		if index != trackedRequest.Index {
			continue
		}

		rv = append(rv, *trackedRequest)
	}
	return rv
}

// GetTrackedRequest returns a pending tracked request.
func (t CustomCookieRequestTracker) GetTrackedRequest(r *http.Request, index string) (*samlsp.TrackedRequest, error) {
	cookie, err := r.Cookie(t.NamePrefix + index)
	if err != nil {
		return nil, err
	}

	trackedRequest, err := t.Codec.Decode(cookie.Value)
	if err != nil {
		return nil, err
	}
	if trackedRequest.Index != index {
		return nil, fmt.Errorf("expected index %q, got %q", index, trackedRequest.Index)
	}
	return trackedRequest, nil
}

func randomBytes(n int) []byte {
	rv := make([]byte, n)
	if _, err := io.ReadFull(saml.RandReader, rv); err != nil {
		panic(err)
	}
	return rv
}
