package toon

import (
	"strings"
)

// KeyFolding controls dotted-key collapsing for nested single-key objects.
type KeyFolding string

const (
	KeyFoldingOff  KeyFolding = "off"
	KeyFoldingSafe KeyFolding = "safe"
)

type encField struct {
	field        Field
	noNestedFold bool
}

func foldObject(obj Object, siblingKeys map[string]struct{}, maxSegments int, pathPrefix string, rootSiblings map[string]struct{}) []encField {
	if len(obj) == 0 || maxSegments < 2 {
		out := make([]encField, len(obj))
		for i, f := range obj {
			out[i] = encField{field: f}
		}
		return out
	}
	out := make([]encField, 0, len(obj))
	for _, field := range obj {
		if folded, ok := foldField(field, siblingKeys, maxSegments, pathPrefix, rootSiblings); ok {
			out = append(out, folded)
			continue
		}
		out = append(out, encField{field: field})
	}
	return out
}

func foldField(field Field, siblingKeys map[string]struct{}, maxSegments int, pathPrefix string, rootSiblings map[string]struct{}) (encField, bool) {
	if !isIdentifierSegment(field.Key) {
		return encField{}, false
	}
	nested, ok := field.Value.(Object)
	if !ok {
		return encField{}, false
	}

	segments := []string{field.Key}
	current := nested

	for len(segments) < maxSegments {
		if len(current) != 1 {
			break
		}
		next := current[0]
		if !isIdentifierSegment(next.Key) {
			break
		}
		candidate := strings.Join(append(segments, next.Key), ".")
		if collides(rootSiblings, pathPrefix, candidate) {
			return encField{}, false
		}
		if _, collision := siblingKeys[candidate]; collision {
			return encField{}, false
		}

		if child, ok := next.Value.(Object); ok && len(child) > 0 {
			segments = append(segments, next.Key)
			current = child
			continue
		}

		segments = append(segments, next.Key)
		return encField{
			field: Field{
				Key:   strings.Join(segments, "."),
				Value: next.Value,
			},
		}, true
	}

	if len(segments) == 1 {
		return encField{}, false
	}

	joined := strings.Join(segments, ".")
	if collides(rootSiblings, pathPrefix, joined) {
		return encField{}, false
	}
	if _, collision := siblingKeys[joined]; collision {
		return encField{}, false
	}
	return encField{
		field:        Field{Key: joined, Value: current},
		noNestedFold: true,
	}, true
}

func collides(rootSiblings map[string]struct{}, pathPrefix, relative string) bool {
	if len(rootSiblings) == 0 {
		return false
	}
	_, hit := rootSiblings[joinPath(pathPrefix, relative)]
	return hit
}

func joinPath(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

func siblingKeySet(obj Object) map[string]struct{} {
	set := make(map[string]struct{}, len(obj))
	for _, f := range obj {
		set[f.Key] = struct{}{}
	}
	return set
}

func maxFoldSegments(opts EncodeOptions) int {
	if opts.KeyFolding != KeyFoldingSafe {
		return 0
	}
	if opts.FlattenDepth != nil {
		if *opts.FlattenDepth < 2 {
			return 0
		}
		return *opts.FlattenDepth
	}
	return -1
}
