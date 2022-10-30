package message

func MentionedText(id, name string) string {
	return `<a href="https://matrix.to/#/` + id + `">` + name + `</a>`
}
