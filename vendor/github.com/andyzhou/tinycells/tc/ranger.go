package tc

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type HttpRange struct {
	Start  int64
	Length int64
}

type Ranger struct {
	http.ResponseWriter
	start  int64
	length int64
	flag   int64
	HttpRange
}

func (w *Ranger) Write(data []byte) (size int, err error) {
	size = len(data)

	if (w.flag+int64(size) <= w.start) || (w.flag >= w.start+w.length) {
		return
	}

	start := w.start - w.flag
	if start < 0 {
		start = 0
	}
	// add flag
	w.flag += int64(size)
	var end int64
	if w.flag <= w.start+w.length {
		end = int64(size)
	} else {
		end = w.start + w.length - (w.flag - int64(size))
	}

	w.ResponseWriter.Write(data[start:end])
	return
}

func (w *Ranger) Process(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Accept-Ranges", "bytes")
	rangeString := req.Header.Get("Range")
	if rangeString == "" {
		return
	}

	// BUG: get total content length (now: 100 for test)
	ranges, err := w.ParseRange(rangeString, 100)
	if err != nil {
		http.Error(res, "Requested Range Not Satisfiable", 416)
		return
	}

	start := ranges[0].Start
	length := ranges[0].Length

	res.Header().Set("Content-Range", w.GetRange(start, start+length-1, -1))
	// res.Header().Set("Content-Length", strconv.FormatInt(length, 10))
	res.WriteHeader(206)
}

// Example:
//   "Range": "bytes=100-200"
//   "Range": "bytes=-50"
//   "Range": "bytes=150-"
//   "Range": "bytes=0-0,-1"
func (w *Ranger) ParseRange(s string, size int64) ([]HttpRange, error) {
	if s == "" {
		return nil, nil // header not present
	}
	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return nil, errors.New("invalid range")
	}
	var ranges []HttpRange
	for _, ra := range strings.Split(s[len(b):], ",") {
		ra = strings.TrimSpace(ra)
		if ra == "" {
			continue
		}
		i := strings.Index(ra, "-")
		if i < 0 {
			return nil, errors.New("invalid range")
		}
		start, end := strings.TrimSpace(ra[:i]), strings.TrimSpace(ra[i+1:])
		var r HttpRange
		if start == "" {
			// If no start is specified, end specifies the
			// range start relative to the end of the file.
			i, err := strconv.ParseInt(end, 10, 64)
			if err != nil {
				return nil, errors.New("invalid range")
			}
			if i > size {
				i = size
			}
			r.Start = size - i
			r.Length = size - r.Start
		} else {
			i, err := strconv.ParseInt(start, 10, 64)
			if err != nil || i >= size || i < 0 {
				return nil, errors.New("invalid range")
			}
			r.Start = i
			if end == "" {
				// If no end is specified, range extends to end of the file.
				r.Length = size - r.Start
			} else {
				i, err := strconv.ParseInt(end, 10, 64)
				if err != nil || r.Start > i {
					return nil, errors.New("invalid range")
				}
				if i >= size {
					i = size - 1
				}
				r.Length = i - r.Start + 1
			}
		}
		ranges = append(ranges, r)
	}
	return ranges, nil
}


// Example:
//   "Content-Range": "bytes 100-200/1000"
//   "Content-Range": "bytes 100-200/*"
func (w *Ranger) GetRange(start, end, total int64) string {
	// unknown total: -1
	if total == -1 {
		return fmt.Sprintf("bytes %d-%d/*", start, end)
	}
	return fmt.Sprintf("bytes %d-%d/%d", start, end, total)
}

