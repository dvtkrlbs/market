package money

import (
	"fmt"
	"strconv"
	"strings"
)

func Parse(value string) (int64, error) {
	value = strings.TrimSpace(value)
	negative := strings.HasPrefix(value, "-")
	value = strings.TrimPrefix(strings.TrimPrefix(value, "+"), "-")
	parts := strings.Split(value, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, fmt.Errorf("amount must look like 12.34 or -12.34")
	}
	if len(parts) == 1 {
		parts = append(parts, "")
	}
	if len(parts[1]) > 2 {
		return 0, fmt.Errorf("amount may have at most two decimal places")
	}
	centsText := parts[1] + strings.Repeat("0", 2-len(parts[1]))
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount %q", value)
	}
	cents, err := strconv.ParseInt(centsText, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount %q", value)
	}
	amount := whole*100 + cents
	if negative {
		amount = -amount
	}
	return amount, nil
}

func Format(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100)
}
