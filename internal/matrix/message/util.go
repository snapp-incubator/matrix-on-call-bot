package message

import "strings"

func FormatCommaSeperatedList(in string) string {
	items := strings.Split(in, ",")

	for i := range items {
		items[i] = strings.TrimSpace(items[i])
	}

	return strings.Join(items, ", ")
}
