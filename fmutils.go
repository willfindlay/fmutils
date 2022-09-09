package fmutils

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Filter keeps the msg fields that are listed in the paths and clears all the rest.
//
// This is a handy wrapper for NestedMask.Filter method.
// If the same paths are used to process multiple proto messages use NestedMask.Filter method directly.
func Filter(msg proto.Message, paths []string) {
	NestedMaskFromPaths(paths).Filter(msg)
}

// Prune clears all the fields listed in paths from the given msg.
//
// This is a handy wrapper for NestedMask.Prune method.
// If the same paths are used to process multiple proto messages use NestedMask.Filter method directly.
func Prune(msg proto.Message, paths []string) {
	NestedMaskFromPaths(paths).Prune(msg)
}

// NestedMask represents a field mask as a recursive map.
type NestedMask map[string]NestedMask

// NestedMaskFromPaths creates an instance of NestedMask for the given paths.
func NestedMaskFromPaths(paths []string) NestedMask {
	mask := make(NestedMask)
	for _, path := range paths {
		curr := mask
		var letters []rune
		for _, letter := range path {
			if letter == '.' {
				if len(letters) == 0 {
					continue
				}

				key := string(letters)
				c, ok := curr[key]
				if !ok {
					c = make(NestedMask)
					curr[key] = c
				}
				curr = c
				letters = nil
				continue
			}
			letters = append(letters, letter)
		}
		if len(letters) != 0 {
			key := string(letters)
			if _, ok := curr[key]; !ok {
				curr[key] = make(NestedMask)
			}
		}
	}

	return mask
}

// Filter keeps the msg fields that are listed in the paths and clears all the rest.
//
// If the mask is empty then all the fields are kept.
// Paths are assumed to be valid and normalized otherwise the function may panic.
// See google.golang.org/protobuf/types/known/fieldmaskpb for details.
func (mask NestedMask) Filter(msg proto.Message) {
	if len(mask) == 0 {
		return
	}

	rft := msg.ProtoReflect()
	rft.Range(func(fd protoreflect.FieldDescriptor, _ protoreflect.Value) bool {
		m, ok := mask[string(fd.Name())]
		if ok {
			if len(m) == 0 {
				return true
			}

			if fd.IsMap() {
				xmap := rft.Get(fd).Map()
				xmap.Range(func(mk protoreflect.MapKey, mv protoreflect.Value) bool {
					if mi, ok := m[mk.String()]; ok {
						if i, ok := mv.Interface().(protoreflect.Message); ok && len(mi) > 0 {
							mi.Filter(i.Interface())
						}
					} else {
						xmap.Clear(mk)
					}

					return true
				})
			} else if fd.IsList() {
				list := rft.Get(fd).List()
				for i := 0; i < list.Len(); i++ {
					m.Filter(list.Get(i).Message().Interface())
				}
			} else if fd.Kind() == protoreflect.MessageKind {
				m.Filter(rft.Get(fd).Message().Interface())
			}
		} else {
			rft.Clear(fd)
		}
		return true
	})
}

// Prune clears all the fields listed in paths from the given msg.
//
// All other fields are kept untouched. If the mask is empty no fields are cleared.
// This operation is the opposite of NestedMask.Filter.
// Paths are assumed to be valid and normalized otherwise the function may panic.
// See google.golang.org/protobuf/types/known/fieldmaskpb for details.
func (mask NestedMask) Prune(msg proto.Message) {
	if len(mask) == 0 {
		return
	}

	rft := msg.ProtoReflect()
	rft.Range(func(fd protoreflect.FieldDescriptor, _ protoreflect.Value) bool {
		m, ok := mask[string(fd.Name())]
		if ok {
			if len(m) == 0 {
				rft.Clear(fd)
				return true
			}

			if fd.IsMap() {
				xmap := rft.Get(fd).Map()
				xmap.Range(func(mk protoreflect.MapKey, mv protoreflect.Value) bool {
					if mi, ok := m[mk.String()]; ok {
						if i, ok := mv.Interface().(protoreflect.Message); ok && len(mi) > 0 {
							mi.Prune(i.Interface())
						} else {
							xmap.Clear(mk)
						}
					}

					return true
				})
			} else if fd.IsList() {
				list := rft.Get(fd).List()
				for i := 0; i < list.Len(); i++ {
					m.Prune(list.Get(i).Message().Interface())
				}
			} else if fd.Kind() == protoreflect.MessageKind {
				m.Prune(rft.Get(fd).Message().Interface())
			}
		}
		return true
	})
}

// WildcardFilter keeps the msg fields that are listed in the paths and clears all the rest.
// Supports wildcard keys in pathnames, specified by an asterisk, "*".
//
// This is a handy wrapper for WildcardNestedMask.Filter method.
// If the same paths are used to process multiple proto messages use WildcardNestedMask.Filter method directly.
func WildcardFilter(msg proto.Message, paths []string) {
	WildcardNestedMaskFromPaths(paths).Filter(msg)
}

// WildcardPrune clears all the fields listed in paths from the given msg.
// Supports wildcard keys in pathnames, specified by an asterisk, "*".
//
// This is a handy wrapper for WildcardNestedMask.Prune method.
// If the same paths are used to process multiple proto messages use WildcardNestedMask.Filter method directly.
func WildcardPrune(msg proto.Message, paths []string) {
	WildcardNestedMaskFromPaths(paths).Prune(msg)
}

// WildcardNestedMask represents a field mask with support for wildcards as a recursive map.
type WildcardNestedMask map[string]WildcardNestedMask

// WildcardNestedMaskFromPaths creates an instance of WildcardNestedMask for the given paths.
func WildcardNestedMaskFromPaths(paths []string) WildcardNestedMask {
	mask := make(WildcardNestedMask)
	for _, path := range paths {
		curr := mask
		var letters []rune
		for _, letter := range path {
			if letter == '.' {
				if len(letters) == 0 {
					continue
				}

				key := string(letters)
				c, ok := curr[key]
				if !ok {
					c = make(WildcardNestedMask)
					curr[key] = c
				}
				curr = c
				letters = nil
				continue
			}
			letters = append(letters, letter)
		}
		if len(letters) != 0 {
			key := string(letters)
			if _, ok := curr[key]; !ok {
				curr[key] = make(WildcardNestedMask)
			}
		}
	}

	return mask
}

func (mask WildcardNestedMask) filterField(
	rft protoreflect.Message,
	fd protoreflect.FieldDescriptor,
	wildcard, pruning bool,
) bool {
	var key string
	if wildcard {
		key = "*"
	} else {
		key = string(fd.Name())
	}

	m, ok := mask[key]
	if ok {
		if len(m) == 0 {
			return true
		}

		if fd.IsMap() {
			xmap := rft.Get(fd).Map()
			xmap.Range(func(mk protoreflect.MapKey, mv protoreflect.Value) bool {
				if mask.filterMapKey(xmap, mk, mv, false) {
					if pruning {
						xmap.Clear(mk)
					}
					return true
				}
				if mask.filterMapKey(xmap, mk, mv, true) {
					if pruning {
						xmap.Clear(mk)
					}
					return true
				}
				if !pruning {
					xmap.Clear(mk)
				}
				return false
			})
		} else if fd.IsList() {
			list := rft.Get(fd).List()
			for i := 0; i < list.Len(); i++ {
				m.Filter(list.Get(i).Message().Interface())
			}
		} else if fd.Kind() == protoreflect.MessageKind {
			m.Filter(rft.Get(fd).Message().Interface())
		}
	}
	return true
}

func (mask WildcardNestedMask) filterMapKey(
	xmap protoreflect.Map,
	mk protoreflect.MapKey,
	mv protoreflect.Value,
	wildcard bool,
) bool {
	var key string
	if wildcard {
		key = "*"
	} else {
		key = mk.String()
	}

	m, ok := mask[key]
	if ok {
		if i, ok := mv.Interface().(protoreflect.Message); ok && len(m) > 0 {
			m.Filter(i.Interface())
		}
	}

	return true
}

// Filter keeps the msg fields that are listed in the paths and clears all the rest.
// Supports wildcard keys in pathnames, specified by an asterisk, "*".
//
// If the mask is empty then all the fields are kept.
// Paths are assumed to be valid and normalized otherwise the function may panic.
// See google.golang.org/protobuf/types/known/fieldmaskpb for details.
func (mask WildcardNestedMask) Filter(msg proto.Message) {
	if len(mask) == 0 {
		return
	}

	rft := msg.ProtoReflect()
	rft.Range(func(fd protoreflect.FieldDescriptor, _ protoreflect.Value) bool {
		if mask.filterField(rft, fd, false, false) {
			return true
		}
		if mask.filterField(rft, fd, true, false) {
			return true
		}
		rft.Clear(fd)
		return false
	})
}

// Prune clears all the fields listed in paths from the given msg.
// Supports wildcard keys in pathnames, specified by an asterisk, "*".
//
// All other fields are kept untouched. If the mask is empty no fields are cleared.
// This operation is the opposite of WildcardNestedMask.Filter.
// Paths are assumed to be valid and normalized otherwise the function may panic.
// See google.golang.org/protobuf/types/known/fieldmaskpb for details.
func (mask WildcardNestedMask) Prune(msg proto.Message) {
	if len(mask) == 0 {
		return
	}

	rft := msg.ProtoReflect()
	rft.Range(func(fd protoreflect.FieldDescriptor, _ protoreflect.Value) bool {
		if mask.filterField(rft, fd, false, true) {
			rft.Clear(fd)
			return true
		}
		if mask.filterField(rft, fd, true, true) {
			rft.Clear(fd)
			return true
		}
		return false
	})
}
