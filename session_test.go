package gosql

import (
	"testing"
)

func TestNewSession(t *testing.T) {
	// s := &Session{ctx: ctx, cluster: c, v: v}
	s := &Session{}
	s.Commit()
}
