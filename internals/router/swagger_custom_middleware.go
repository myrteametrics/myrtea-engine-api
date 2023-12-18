package router

import (
	"net/http"
	"strings"

	_ "github.com/myrteametrics/myrtea-engine-api/v5/docs" // docs is generated by Swag CL
	"github.com/spf13/viper"
)

func SwaggerUICustomizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.RequestURI, "/swagger/") {
			rec := newResponseRecorder(w)
			next.ServeHTTP(rec, r)

			if strings.Contains(rec.Header().Get("Content-Type"), "text/html") {
				modifiedHTML := modifySwaggerHTML(rec.body.String())
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte(modifiedHTML))
			}
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func modifySwaggerHTML(html string) string {
	topbarColor := viper.GetString("SWAGGER_TOPBAR_COLOR")
	topbarTitle := viper.GetString("SWAGGER_TOPBAR_TITLE")

	script := `
	<script>
	window.addEventListener('DOMContentLoaded', function() {
		const color = '` + topbarColor + `';
		const title = '` + topbarTitle + `';

		const customStyle = document.createElement('style');
		customStyle.textContent = "div.topbar { background-color: " + color + " !important; }"

		const customHeader = document.createElement('div');
		customHeader.style.position = 'sticky';
		customHeader.style.top = '0';
		customHeader.style.zIndex = '2';
		customHeader.style.textAlign = 'center';
		customHeader.style.fontSize = '1.5em';
		customHeader.style.fontWeight = '700';
		customHeader.style.fontFamily = 'Titillium Web,sans-serif';
		customHeader.style.color = '#ffffff';
		customHeader.style.backgroundColor = color;
		customHeader.textContent = title;

		document.title = title;
		document.head.appendChild(customStyle);
		const targetElement = document.querySelector('body');
		if (targetElement) {
			targetElement.prepend(customHeader);
		}
	});
	</script>
	`

	// position to inject the script
	pos := strings.LastIndex(html, "</body>")
	if pos == -1 {
		return html
	}

	return html[:pos] + script + html[pos:]
}

// responseRecorder is a struct that is used to capture the response
type responseRecorder struct {
	http.ResponseWriter
	body *strings.Builder
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{ResponseWriter: w, body: new(strings.Builder)}
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
