package matrix

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/matrix-org/gomatrix"
	"github.com/pkg/errors"
	"github.com/snapp-incubator/matrix-on-call-bot/internal/model"
)

type Head string

const (
	incoming string = "incoming"
	outgoing string = "outgoing"

	minCommandLength int = 1

	CreateFollowUp          Head = "!followup" // !followup <category: incoming|outgoing> <initiator> <description>
	minCreateFollowUpLength int  = 4

	ListFollowUp Head = "!listfollowups" // !listfollowups

	ResolveFollowUp          Head = "!resolvefollowup" // !resolvefollowup <id>
	minResolveFollowUpLength int  = 2

	Report Head = "!report" // !report

	Help = "!help" // !help
)

var (
	ErrInvalidCommand = errors.New("invalid command")
	ErrUnknownCommand = errors.New("unknown command")
	ErrInvalidBody    = errors.New("invalid body")
	ErrInvalidType    = errors.New("invalid type")

	// Regexp is a compiled regular expression that can extract data in a message containing people mentioning (like:
	// @ahmad.anvari:snapp.cab).
	Regexp = regexp.MustCompile(`<a href="https://matrix.to/#/(.*?)">(.*?)</a>`)
)

type Command interface {
	Match(message string) bool
	Handle(event *gomatrix.Event) error
}

//nolint:cyclop
func (b *Bot) Handle(event *gomatrix.Event) error {
	raw, ok := event.Content["body"].(string)
	if !ok {
		return ErrInvalidBody
	}

	parts := strings.Split(raw, " ")

	if len(parts) < minCommandLength {
		return ErrInvalidCommand
	}

	if parts[0][0] != '!' {
		return nil
	}

	switch Head(parts[0]) {
	case CreateFollowUp:
		return b.createFollowUp(event, parts)
	case ListFollowUp:
		return b.listFollowUps(event)
	case ResolveFollowUp:
		return b.resolveFollowUp(event, parts)
	case Report:
		return b.report(event, parts)
	case Help:
		return b.help(event)
	default:
		return ErrUnknownCommand
	}
}

func (b *Bot) createFollowUp(event *gomatrix.Event, parts []string) error {
	if len(parts) < minCreateFollowUpLength {
		return ErrInvalidCommand
	}

	active, err := b.shiftRepo.Active(event.RoomID)
	if err != nil {
		return errors.Wrap(err, "error getting active shifts")
	}

	if len(active) < 1 {
		if _, err := b.cli.SendText(event.RoomID, NoActiveShiftOngoing); err != nil {
			return errors.Wrap(err, "error sending no active shift message")
		}

		return nil
	}

	shiftID := active[0].ID
	sender := event.Sender
	category := b.followUpCategory(parts[1])
	initiator := parts[2]
	description := ""

	for _, descPart := range parts[3:] {
		description += descPart + " "
	}

	description = strings.TrimSpace(description)

	followUp := model.FollowUp{
		ShiftID:     shiftID,
		Sender:      sender,
		Initiator:   initiator,
		Description: description,
		Done:        false,
		Category:    category,
	}

	if err := b.followUpRepo.Create(&followUp); err != nil {
		return errors.Wrap(err, "error saving follow up")
	}

	m := fmt.Sprintf(FollowUpCreated, ListFollowUp, ResolveFollowUp, followUp.ID)

	if _, err := b.cli.SendFormattedText(event.RoomID, "", m); err != nil {
		return errors.Wrap(err, "error sending create follow up response")
	}

	return nil
}

func (b *Bot) listFollowUps(event *gomatrix.Event) error {
	active, err := b.shiftRepo.Active(event.RoomID)
	if err != nil {
		return errors.Wrap(err, "error getting active shifts")
	}

	if len(active) < 1 {
		if _, err := b.cli.SendText(event.RoomID, NoActiveShiftOngoing); err != nil {
			return errors.Wrap(err, "error sending no active shift message")
		}

		return nil
	}

	shiftID := active[0].ID

	items, err := b.followUpRepo.Get(shiftID)
	if err != nil {
		return errors.Wrap(err, "error getting follow ups")
	}

	message := ""

	for _, item := range items {
		emoji := "⭕️"

		if item.Done {
			emoji = "✅"
		}

		message += fmt.Sprintf(FollowUpItem,
			emoji, item.ID, item.Category,
			item.Initiator, item.Description, item.CreatedAt.Local().Format(time.RFC850))
	}

	message = fmt.Sprintf(FollowUpList, message)

	if _, err := b.cli.SendFormattedText(event.RoomID, "", message); err != nil {
		return errors.Wrap(err, "error sending list of follow ups")
	}

	return nil
}

func (b *Bot) resolveFollowUp(event *gomatrix.Event, parts []string) error {
	if len(parts) < minResolveFollowUpLength {
		return ErrInvalidCommand
	}

	idStr := parts[1]

	followUpID, err := strconv.Atoi(idStr)
	if err != nil {
		return errors.Wrap(err, "error converting follow up id to int")
	}

	if err := b.followUpRepo.Update(&model.FollowUp{ID: followUpID, Done: true}); err != nil {
		return errors.Wrap(err, "error updating follow up")
	}

	if _, err := b.cli.SendFormattedText(event.RoomID, "", fmt.Sprintf(FollowUpResolved, followUpID)); err != nil {
		return errors.Wrap(err, "error sending follow up resolved message")
	}

	return nil
}

func (b *Bot) help(event *gomatrix.Event) error {
	_, err := b.cli.SendFormattedText(event.RoomID, "", HelpList)
	if err != nil {
		return errors.Wrap(err, "error sending help response")
	}

	return nil
}

func (b *Bot) followUpCategory(in string) string {
	switch in {
	case "in", incoming:
		return incoming
	case "out", outgoing:
		return outgoing
	default:
		return incoming
	}
}

func (b *Bot) mentionedText(id, name string) string {
	return `<a href="https://matrix.to/#/` + id + `">` + name + `</a>`
}

type ShiftReportTemplate struct {
	Items []ShiftReportItemTemplate
	From  string
	To    string
}

type ShiftReportItemTemplate struct {
	HolderID   string
	WorkingDay int
	Holiday    int
}

// nolint: funlen,gocognit, cyclop
func (b *Bot) report(event *gomatrix.Event, parts []string) error {
	// Handling custom time range report
	//nolint: varnamelen
	var from, to time.Time

	var err error

	switch {
	case len(parts) == 1: // ["!report"]
		to = time.Now()
		from = time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, time.Local)
	case len(parts) == 3 && strings.EqualFold(parts[1], "from"): // ["!report", "FROM", "2022-10-17"]
		to = time.Now()

		from, err = time.Parse(time.RFC3339, parts[2]+"T00:00:00Z")
		if err != nil {
			if _, err = b.cli.SendText(event.RoomID, fmt.Sprintf(InvalidReportCommandWithError, err.Error())); err != nil {
				return errors.Wrap(err, "error sending invalid report command message")
			}

			return nil
		}
	case len(parts) == 5 && strings.EqualFold(parts[1], "from") &&
		strings.EqualFold(parts[3], "to"): // ["!report", "FROM", "2022-10-17", "TO", "2022-10-21"]
		from, err = time.Parse(time.RFC3339, parts[2]+"T00:00:00Z")
		if err != nil {
			if _, err = b.cli.SendText(event.RoomID, fmt.Sprintf(InvalidReportCommandWithError, err.Error())); err != nil {
				return errors.Wrap(err, "error sending invalid report command message")
			}

			return nil
		}

		to, err = time.Parse(time.RFC3339, parts[4]+"T00:00:00Z")
		if err != nil {
			if _, err = b.cli.SendText(event.RoomID, fmt.Sprintf(InvalidReportCommandWithError, err.Error())); err != nil {
				return errors.Wrap(err, "error sending invalid report command message")
			}

			return nil
		}
	default: // Not valid format
		if _, err = b.cli.SendText(event.RoomID, InvalidReportCommand); err != nil {
			return errors.Wrap(err, "error sending invalid report command message")
		}

		return nil
	}

	shifts, err := b.shiftRepo.Report(event.RoomID, from, to)
	if err != nil {
		return errors.Wrap(err, "error in getting shifts from the db")
	}

	shiftsRep := make([]ShiftReportItemTemplate, 0, len(shifts))
	results := make(map[string]ShiftReportItemTemplate)

	for _, shift := range shifts {
		var temp ShiftReportItemTemplate

		var ok bool

		if temp, ok = results[shift.Holders]; !ok {
			temp = ShiftReportItemTemplate{
				HolderID:   shift.Holders,
				WorkingDay: 0,
				Holiday:    0,
			}
		}

		if shift.EndTime == nil {
			shift.EndTime = &to
		}

		if from.After(shift.StartTime) {
			shift.StartTime = from
		}

		if to.Before(*shift.EndTime) {
			shift.EndTime = &to
		}

		wd, hd := dateDiff(shift.StartTime, *shift.EndTime)
		temp.WorkingDay += wd
		temp.Holiday += hd
		results[shift.Holders] = temp
	}

	for _, result := range results {
		displayName, err := b.cli.GetDisplayName(result.HolderID)
		if err != nil {
			return errors.Wrap(err, "error getting the display name of the event sender")
		}

		shiftsRep = append(shiftsRep, ShiftReportItemTemplate{
			HolderID:   b.mentionedText(result.HolderID, displayName.DisplayName),
			WorkingDay: result.WorkingDay,
			Holiday:    result.Holiday,
		})
	}

	tmp := ShiftReportTemplate{
		Items: shiftsRep,
		From:  from.Format(time.Stamp),
		To:    to.Format(time.Stamp),
	}

	var buf bytes.Buffer

	err = reportTemplate.Execute(&buf, tmp)
	if err != nil {
		return errors.Wrap(err, "error in executing the template with parameter")
	}

	if _, err := b.cli.SendFormattedText(event.RoomID, "monthly report", buf.String()); err != nil {
		return errors.Wrap(err, "error in sending monthly report")
	}

	return nil
}

const (
	weekDays = 7
	dayHours = 24
)

func dateDiff(start, end time.Time) (int, int) {
	var normalDays, holidays int
	// List of days those are holidays during the week. For example Thursday and Friday is holiday in my country
	weekHolidays := []time.Weekday{
		time.Thursday,
		time.Friday,
	}

	// Calculate number of days between start and end
	diffDays := end.Sub(start).Hours()/dayHours + 1
	// Calculate number of complete weeks between start and end
	fullWeeks := math.Floor(diffDays / weekDays)

	// Each full weeks have the number holidays during it
	fullWeeksHolidays := int(fullWeeks) * len(weekHolidays)

	if uint(diffDays)%weekDays == 0 {
		holidays = fullWeeksHolidays
	} else {
		// nEnd is the end of the last full week
		nEnd := start.Add(time.Duration(fullWeeks) * weekDays * dayHours * time.Hour)
		counter := 0

		// Calculate number of holidays during nEnd to end
		for _, weekHolidayDay := range weekHolidays {
			if nEnd.Weekday() <= weekHolidayDay && end.Weekday() >= weekHolidayDay {
				counter++
			}
		}

		holidays = fullWeeksHolidays + counter
	}

	normalDays = int(diffDays) - holidays

	return normalDays, holidays
}
