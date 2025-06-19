package redirect_test

import (
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-urlshortner/internal/http-server/handlers/url/redirect"
	"go-urlshortner/internal/http-server/handlers/url/redirect/mocks"
	"go-urlshortner/internal/lib/api"
	"go-urlshortner/internal/lib/logger/handlers/slogdiscard"
	"net/http/httptest"
	"testing"
)

func TestRedirectHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{
			name:  "Success",
			alias: "test_alias",
			url:   "https://google.com",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			urlGetterMock := mocks.NewMockURLGetter(t)

			if tc.respError == "" || tc.mockError != nil {
				urlGetterMock.On("GetURL", tc.alias).Return(tc.url, tc.mockError).Once()
			}

			r := chi.NewRouter()
			r.Get("/{alias}", redirect.New(slogdiscard.NewDiscardLogger(), urlGetterMock))

			ts := httptest.NewServer(r)
			defer ts.Close()

			redirectedToURL, err := api.GetRedirect(ts.URL + "/" + tc.alias)
			require.NoError(t, err)

			assert.Equal(t, tc.url, redirectedToURL)
		})
	}
}
