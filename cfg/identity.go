package cfg

import (
	"github.com/ubclaunchpad/inertia/cfg/internal/identity"
)

// ident wraps an array of whatever into identity.Identifier. Why does this
// exist? Because you can't have `Set(v string, ids ...Identifier)` and:
//
//    profiles := []*Profile{ ... }
//    identity.Set(v, profiles...)
//
// You get:
//
//    cannot use p.Profiles (type []*Profile) as type []identity.Identifier in argument to identity.Set
//
// Even though this works:
//
//    profiles := []*Profile{ ... }
//    identity.Set(v, profiles[0], profiles[1])
//
// tl;dr generics please? :(
func ident(vals interface{}) []identity.Identifier {
	var ids []identity.Identifier
	switch impl := vals.(type) {
	case []*Project:
		ids = make([]identity.Identifier, len(impl))
		for i, v := range impl {
			ids[i] = v
		}
	case []*Remote:
		ids = make([]identity.Identifier, len(impl))
		for i, v := range impl {
			ids[i] = v
		}
	}
	return ids
}
