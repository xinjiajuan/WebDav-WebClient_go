package Object

import (
	"strconv"
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
