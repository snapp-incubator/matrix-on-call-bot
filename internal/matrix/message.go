package matrix

import (
	"text/template"
)

//nolint:gochecknoglobals
var reportTemplate = template.Must(template.New("tmpl").Parse(ReportMessage))

//nolint:lll
const (
	ShiftStarted         = `%s shift started at %s.`
	ShiftItem            = "<li>%s <b>Start time</b>: %s | <b>End time</b>: %s</li> | <b>Holders</b>: %s | <b>id</b>: %d"
	ShiftList            = `<ol>%s</ol>`
	ShiftEndFormatted    = "Shift with id: <b>%d</b> ended. Good job! :)"
	ShiftEnd             = "Shift with id %d ended."
	InvalidShiftStart    = "Please mention the on call people."
	ActiveShiftOngoing   = "There's an active shift still in progress. You can't start a new one."
	NoActiveShiftOngoing = "There's no active shift. Create one first."
	FollowUpCreated      = "Follow up created. List all follow ups with %s or mark this follow up as resolved by %s %d"
	FollowUpItem         = "<li>%s <b>id</b>: %d | <b>Category</b>: %s</li> | " +
		"<b>Initiator</b>: %s</li> | <b>Description</b>: %s | <b>Created at</b>: %s"
	FollowUpList     = `<ol>%s</ol>`
	FollowUpResolved = "Follow up with id: <b>%d</b>, marked as resolved."
	HelpList         = `
<h2>Shift commands:</h2>
<ul>
<li>!startshift &lt;mentioned oncalls&gt; <b>=&gt;</b> start a new shift for the sender of the message or if anyone is mentioned, start shift for the mentioned people</li>
<li>!listshifts <b>=&gt;</b> list all shifts</li>
<li>!endshift &lt;shift id&gt; <b>=&gt;</b> end a shift</li>
</ul>
<br>
<h2>Follow up commands:</h2>
<ul>
<li>!followup &lt;category: incoming|outgoing&gt; &lt;initiator&gt; &lt;description&gt; <b>=&gt;</b> create a new follow up</li>
<li>!listfollowups <b>=&gt;</b> list all follow ups</li>
<li>!resolvefollowup <id> <b>=&gt;</b> resolve a follow up</li>
<li>!report<b>=&gt;</b> Report current room on-call days for this month</li>
</ul>
`
	ReportMessage = `
<ul>
{{range $item := .}}
    <li> {{$item.HolderID}}
		<ul>
			<li>Working day: {{$item.WorkingDay}}</li>
			<li>Holiday: {{$item.Holiday}}</li>
		</ul>
	</li>
{{end}}
</ul>
`
)
