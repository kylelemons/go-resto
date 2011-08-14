package rest

import (
	"sort"
	"strconv"
	"strings"
)

type MediaType struct {
	Type string
	SubType string
	Quality float64
	Params map[string]string
	Index int
}

// Match returns true if the known media type matches the req'd one.
// If the type and subtype match, the parameters are checked:
//   If req has a parameter known doesn't, they don't match
//   If both have a parameter whose values don't match, they don't match
//   Otherwise, they match
func (req *MediaType) Match(known *MediaType) bool {
	if req.Type != "*" && req.Type != known.Type {
		return false
	}
	if req.SubType != "*" && req.SubType != known.SubType {
		return false
	}
	for k, v := range req.Params {
		if ov, ok := known.Params[k]; !ok || ov != v {
			return false
		}
	}
	return true
}

// Copy performs a deep copy of mt.
func (mt *MediaType) Copy() *MediaType {
	cp := new(MediaType)
	*cp = *mt
	cp.Params = make(map[string]string, len(mt.Params))
	for k, v := range mt.Params {
		cp.Params[k] = v
	}
	return cp
}

func (mt *MediaType) String() string {
	pieces := []string{
		mt.Type+"/"+mt.SubType,
	}
	for k, v := range mt.Params {
		pieces = append(pieces, k + "=" + v)
	}
	if mt.Quality != 1.0 {
		pieces = append(pieces, "q="+strconv.Ftoa64(mt.Quality, 'g', -1))
	}
	return strings.Join(pieces, ";")
}

// Compare compares the two media types for their relative quality and returns:
//   <0 if a is higher quality than b
//   =0 if a is the same quality as b (or if a and b's types don't match)
//   >0 if a is lower quality than b
func (a *MediaType) Compare(b *MediaType) int {
	switch diff := a.Quality - b.Quality; {
	case diff < 0: return +1
	case diff > 0: return -1
	}
	if a.Type == b.Type {
		if a.SubType == b.SubType {
			return len(b.Params) - len(a.Params)
		}
		switch {
		case a.SubType == "*": return +1
		case b.SubType == "*": return -1
		}
		return a.Index - b.Index
	}
	switch {
	case a.Type == "*": return +1
	case b.Type == "*": return -1
	}
	return a.Index - a.Index
}

// MediaTypeList is a sort.Sort-able list of media types.
type MediaTypeList []*MediaType
func (l MediaTypeList) Len() int { return len(l) }
func (l MediaTypeList) Less(i, j int) bool { return l[i].Compare(l[j]) < 0 }
func (l MediaTypeList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

func (l MediaTypeList) String() string {
	strs := make([]string, 0, len(l))
	for _, mt := range l {
		strs = append(strs, mt.String())
	}
	return strings.Join(strs, ", ")
}

// Sort sorts the list with highest quality first
func (l MediaTypeList) Sort() {
	sort.Sort(l)
}

// Filter returns a sorted list of the remaining media types in highest- to
// lowest-quality order after filtering the requested media types by the known
// media types.
//
// Warning: This function has abuse potential; prefer Choose over Filter
// when the requested list comes from a user.
func (known MediaTypeList) Filter(requested MediaTypeList) MediaTypeList {
	var out MediaTypeList
	for _, k := range known {
		for _, req := range requested {
			if req.Match(k) {
				mt := k.Copy()
				mt.Index = len(out)
				mt.Quality = k.Quality * req.Quality
				out = append(out, mt)
			}
		}
	}

	sort.Sort(out)
	return out
}

// Choose returns what would be the first result of Filter given the same
// arguments; that is, the highest quality media type from requested that is in
// the known media type list.  Choose returns nil if no matching media types
// are found.
func (known MediaTypeList) Choose(requested MediaTypeList) *MediaType {
	var mt *MediaType
	var temp MediaType
	for _, k := range known {
		for _, req := range requested {
			if req.Match(k) {
				temp = *k
				temp.Quality = k.Quality * req.Quality
				if mt == nil || temp.Compare(mt) < 0 {
					mt = temp.Copy()
				}
			}
		}
	}

	return mt
}

// ParseMediaTypes parses a list of media types into a MediaTypeList.
func ParseMediaTypes(accept []string) (types MediaTypeList) {
	var toks []string
	for _, tok := range accept {
		toks = append(toks, strings.Split(tok, ",")...)
	}

	for _, tok := range toks {
		pieces := strings.Split(tok, ";")
		for i := range pieces {
			pieces[i] = strings.TrimSpace(pieces[i])
		}

		ms := strings.Split(pieces[0]+"/*", "/")
		media, subtype := ms[0], ms[1]

		q := 1.0
		pmap := map[string]string{}
		for _, param := range pieces[1:] {
			if strings.HasPrefix(param, "q=") {
				q, _ = strconv.Atof64(param[2:])
				continue
			}
			kv := strings.Split(param+"=true", "=")
			pmap[kv[0]] = kv[1]
		}

		types = append(types, &MediaType{
			Type: media,
			SubType: subtype,
			Quality: q,
			Params: pmap,
			Index: len(types),
		})
	}
	return
}
