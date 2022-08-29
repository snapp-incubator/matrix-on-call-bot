package matrix

//nolint:lll
const (
	ShiftStarted         = "Shift started at %s. Holders are <b>%s</b>."
	ShiftItem            = "<li>%s <b>Start time</b>: %s | <b>End time</b>: %s</li> | <b>Holders</b>: %s | <b>id</b>: %d"
	ShiftList            = `<ol>%s</ol>`
	ShiftEnd             = "Shift with id: <b>%d</b> ended. Good job! :)"
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
<li>!startshift &lt;comma separated oncall names&gt; <b>=&gt;</b> start a new shift</li>
<li>!listshifts <b>=&gt;</b> list all shifts</li>
<li>!endshift &lt;shift id&gt; <b>=&gt;</b> end a shift</li>
</ul>
<br>
<h2>Follow up commands:</h2>
<ul>
<li>!followup &lt;category: incoming|outgoing&gt; &lt;initiator&gt; &lt;description&gt; <b>=&gt;</b> create a new follow up</li>
<li>!listfollowups <b>=&gt;</b> list all follow ups</li>
<li>!resolvefollowup <id> <b>=&gt;</b> resolve a follow up</li>
</ul>
`
)
