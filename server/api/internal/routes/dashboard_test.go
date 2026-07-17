package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/tom-draper/api-analytics/server/api/internal/config"
)

// requestsRecorderFor drives getRequestsHandler far enough to see how the page
// query is handled. The handler rejects a bad page before touching the
// database, so a nil DB is only reached if validation wrongly lets it through,
// which would panic and fail the test loudly.
func requestsRecorderFor(t *testing.T, rawQuery string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/requests/some-user-id?"+rawQuery, nil)
	c.Params = gin.Params{{Key: "userID", Value: "some-user-id"}}

	getRequestsHandler(nil, &config.Config{PageSize: 250_000, MaxLoad: 1_000_000})(c)
	return recorder
}

// An unparseable page must not reach the database. strconv.Atoi returns 0 on
// failure and page 0 means "load every page", so falling through would turn a
// typo into the most expensive query the endpoint can run.
func TestGetRequestsHandlerRejectsUnparseablePage(t *testing.T) {
	for _, rawQuery := range []string{
		"page=abc",
		"page=1.5",
		"page=",
		"page=%20",
		"page=one",
		"page=1e3",
		"page=0x10",
		"page=" + strings.Repeat("9", 40), // overflows int
	} {
		t.Run(rawQuery, func(t *testing.T) {
			// An empty value is treated as absent, so it keeps the default page.
			if rawQuery == "page=" {
				t.Skip("empty page is absent, covered by TestGetRequestsHandlerAbsentPage")
			}

			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("%s reached the database instead of being rejected: %v", rawQuery, r)
				}
			}()

			recorder := requestsRecorderFor(t, rawQuery)
			if recorder.Code != http.StatusBadRequest {
				t.Errorf("status = %d, expected %d for %s", recorder.Code, http.StatusBadRequest, rawQuery)
			}
			if !strings.Contains(recorder.Body.String(), "Invalid page number") {
				t.Errorf("body = %q, expected an invalid page message", recorder.Body.String())
			}
		})
	}
}

// A valid page must still be accepted and reach the database layer.
func TestGetRequestsHandlerAcceptsValidPage(t *testing.T) {
	for _, rawQuery := range []string{"page=1", "page=3", "page=0"} {
		t.Run(rawQuery, func(t *testing.T) {
			defer func() {
				// Reaching the nil DB proves the page passed validation.
				_ = recover()
			}()

			recorder := requestsRecorderFor(t, rawQuery)
			if recorder.Code == http.StatusBadRequest {
				t.Errorf("%s was rejected, expected it to be accepted", rawQuery)
			}
		})
	}
}

func TestGetRequestsHandlerAbsentPage(t *testing.T) {
	defer func() {
		_ = recover()
	}()

	recorder := requestsRecorderFor(t, "")
	if recorder.Code == http.StatusBadRequest {
		t.Error("an absent page was rejected, expected it to default to page 1")
	}
}
