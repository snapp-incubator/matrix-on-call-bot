package message

const (
	ShiftStarted       = "Shift started at %s. Holders are <b>%s</b>."
	ShiftItem          = "<li>%s <b>Start time</b>: %s | <b>End time</b>: %s</li> | <b>Holders</b>: %s | <b>id</b>: %d"
	ShiftList          = `<ol>%s</ol>`
	ShiftEnd           = "Shift with id: <b>%d</b> ended. Good job! :)"
	ActiveShiftOngoing = "There's an active shift still in progress. You can't start a new one."
	ShiftEndFormatted  = "Shift with id: <b>%d</b> ended. Good job! :)"
	InvalidShiftStart  = "Please mention the on call people."
)
