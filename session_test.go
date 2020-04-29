package gosql

import (
	"testing"
)

func TestNewSession(t *testing.T) {
	Debug = true
	// s := &Session{ctx: ctx, cluster: c, v: v}
	s := &Session{}
	s.Master()
	s.Commit()
	s.Rollback()
}
