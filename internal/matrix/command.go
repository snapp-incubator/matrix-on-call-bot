package matrix

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/matrix-org/gomatrix"

	"github.com/snapp-incubator/matrix-on-call-bot/internal/model"
)

type Head string

const (
	incoming string = "incoming"
	outgoing string = "outgoing"

	ListShift        Head = "!listshifts" // !listshifts
	minCommandLength int  = 1

	CreateShift          Head = "!startshift" // !startshift <comma separated oncall names>
	minCreateShiftLength int  = 2

	EndShift          Head = "!endshift" // !endshift <shift id>
	minEndShiftLength int  = 2

	CreateFollowUp          Head = "!followup" // !followup <category: incoming|outgoing> <initiator> <description>
	minCreateFollowUpLength int  = 4

	ListFollowUp Head = "!listfollowups" // !listfollowups

	ResolveFollowUp          Head = "!resolvefollowup" // !resolvefollowup <id>
	minResolveFollowUpLength int  = 2

	Help = "!help" // !help
)

var (
	ErrInvalidCommand = errors.New("invalid command")
	ErrUnknownCommand = errors.New("unknown command")
	ErrInvalidBody    = errors.New("invalid body")
)

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

	switch Head(parts[0]) {
	case CreateShift:
		return b.createShift(event, parts)
	case EndShift:
		return b.endShift(event, parts)
	case ListShift:
		return b.listShifts(event)
	case CreateFollowUp:
		return b.createFollowUp(event, parts)
	case ListFollowUp:
		return b.listFollowUps(event)
	case ResolveFollowUp:
		return b.resolveFollowUp(event, parts)
	case Help:
		return b.help(event)
	default:
		return ErrUnknownCommand
	}
}

func (b *Bot) createShift(event *gomatrix.Event, parts []string) error {
	if len(parts) < minCreateShiftLength {
		return ErrInvalidCommand
	}

	active, err := b.shiftRepo.Active(event.RoomID)
	if err != nil {
		return errors.Wrap(err, "error getting active shifts")
	}

	if len(active) > 0 {
		if _, err := b.cli.SendText(event.RoomID, ActiveShiftOngoing); err != nil {
			return errors.Wrap(err, "error sending active shift message")
		}

		return nil
	}

	now := time.Now()

	if err := b.shiftRepo.Create(&model.Shift{
		RoomID:    event.RoomID,
		Sender:    event.Sender,
		Holders:   parts[1],
		StartTime: now,
		EndTime:   nil,
	}); err != nil {
		return errors.Wrap(err, "error saving shift")
	}

	_, err = b.cli.SendFormattedText(event.RoomID, "",
		fmt.Sprintf(ShiftStarted, now.Local().Format(time.RFC850), b.commaSeparatedToList(parts[1])))
	if err != nil {
		return errors.Wrap(err, "error sending shift created message")
	}

	return nil
}

func (b *Bot) endShift(event *gomatrix.Event, parts []string) error {
	if len(parts) < minEndShiftLength {
		return ErrInvalidCommand
	}

	shiftID, err := strconv.Atoi(parts[1])
	if err != nil {
		return errors.Wrap(err, "invalid shift id")
	}

	now := time.Now()

	if err := b.shiftRepo.Update(&model.Shift{
		ID:      shiftID,
		EndTime: &now,
	}); err != nil {
		return errors.Wrap(err, "error updating shift")
	}

	if _, err := b.cli.SendFormattedText(event.RoomID, "", fmt.Sprintf(ShiftEnd, shiftID)); err != nil {
		return errors.Wrap(err, "error sending shift end message")
	}

	return nil
}

func (b *Bot) listShifts(event *gomatrix.Event) error {
	list, err := b.shiftRepo.Get(event.RoomID)
	if err != nil {
		return errors.Wrap(err, "error getting shifts")
	}

	message := ""

	for _, item := range list {
		end := "-"
		emoji := "🟢"

		if item.EndTime != nil {
			end = item.EndTime.Local().Format(time.RFC850)
			emoji = "⚪️"
		}

		message += fmt.Sprintf(ShiftItem, emoji,
			item.StartTime.Local().Format(time.RFC850), end, b.commaSeparatedToList(item.Holders), item.ID)
	}

	message = fmt.Sprintf(ShiftList, message)

	if _, err := b.cli.SendFormattedText(event.RoomID, "", message); err != nil {
		return errors.Wrap(err, "error sending shifts list")
	}

	return nil
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

func (b *Bot) commaSeparatedToList(in string) string {
	list := strings.Split(in, ",")

	out := "["

	for i := range list {
		if i == len(list)-1 {
			out += list[i]
		} else {
			out += list[i] + ", "
		}
	}

	return out + "]"
}
