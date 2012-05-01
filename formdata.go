package formdata

import (
	"github.com/sunfmin/reflectutils"
	"net/http"
	"strings"
)

func Unmarshal(r *http.Request, v interface{}, prefix string) (err error) {

	if r.Form == nil && r.MultipartForm == nil {
		r.ParseMultipartForm(32 << 20)
	}

	var vals map[string][]string
	if r.MultipartForm != nil {
		vals = r.MultipartForm.Value
	} else if r.Form != nil {
		vals = map[string][]string(r.Form)
	}

	for fk, fv := range vals {
		key, skip := skipOrTrim(prefix, fk)
		if skip {
			continue
		}
		for _, velem := range fv {
			reflectutils.Set(v, key, velem)
		}
	}

	if r.MultipartForm != nil {

		for filek, filev := range r.MultipartForm.File {
			key, skip := skipOrTrim(prefix, filek)
			if skip {
				continue
			}
			for _, velem := range filev {
				reflectutils.Set(v, key, velem)
			}
		}
		return
	}

	return
}

func skipOrTrim(prefix string, key string) (stripped string, skip bool) {
	if prefix == "" {
		stripped = key
		skip = false
		return
	}
	if strings.Index(key, prefix) != 0 {
		skip = true
		return
	}
	stripped = key[len(prefix):]
	skip = false
	return
}
