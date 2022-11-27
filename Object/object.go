package Object

import (
	"net"
	"net/http"
	"strconv"
	"strings"
)

func getObjectSizeSuitableUnit(size int64) string {
	if size < 1048576 {
		unit := float64(size) / 1024
		return strconv.FormatFloat(unit, 'f', 2, 64) + " KiB"
	} else if size < 1073741824 {
		unit := float64(size) / 1048576
		return strconv.FormatFloat(unit, 'f', 2, 64) + " MiB"
	} else if size < 1099511627776 {
		unit := float64(size) / 1073741824
		return strconv.FormatFloat(unit, 'f', 2, 64) + " GiB"
	} else {
		unit := float64(size) / 1099511627776
		return strconv.FormatFloat(unit, 'f', 2, 64) + " TiB"
	}
}

// GetIP returns request real ip.
func GetIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if net.ParseIP(ip) != nil {
		return ip
	}
	ip = r.Header.Get("X-Forward-For")
	for _, i := range strings.Split(ip, ",") {
		if net.ParseIP(i) != nil {
			return i
		}
	}
	return r.RemoteAddr
}
