package math

import (
	"fmt"
	"strconv"
	"strings"
)

func FormatCurrency(value float64) string {
	intPart := int(value)
	decimalPart := int((value - float64(intPart)) * 100)
	intString := AddCommasToInteger(intPart)
	return fmt.Sprintf("Rp. %s,%02d", intString, decimalPart)
}

func AddCommasToInteger(value int) string {
	strValue := strconv.Itoa(value)
	var parts []string
	for i := len(strValue); i > 0; i -= 3 {
		start := i - 3
		if start < 0 {
			start = 0
		}
		parts = append([]string{strValue[start:i]}, parts...)
	}
	return strings.Join(parts, ",")
}