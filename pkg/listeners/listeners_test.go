package listeners

import (
	"net/url"
	"testing"
)

func TestListenerURIs(t *testing.T) {
	for _, l := range Listeners {
		_, err := url.ParseRequestURI(l.Path)
		if err != nil {
			t.Errorf("%s is not a valid url", l.Path)
		}
	}
}
