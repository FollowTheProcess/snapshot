package snapshot_test

import (
	"testing"

	"github.com/FollowTheProcess/snapshot"
)

func TestHello(t *testing.T) {
	got := snapshot.Hello()
	want := "Hello snapshot"

	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}
}
