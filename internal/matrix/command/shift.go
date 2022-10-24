package command

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/matrix-org/gomatrix"
	"github.com/pkg/errors"
	"github.com/snapp-incubator/matrix-on-call-bot/internal/matrix"
	"github.com/snapp-incubator/matrix-on-call-bot/internal/matrix/message"
	"github.com/snapp-incubator/matrix-on-call-bot/internal/model"
)

const (
	// check if command starts with
	// !startshift or
	// !endshift or
	// !listshifts.
	startShiftParseRegexString = "^!startshift|^!endshift|^!listshifts"

	minCreateShiftLength int = 2

	minEndShiftLength int = 2
)

type Shift struct {
	parseRegex *regexp.Regexp
	shiftRepo  model.ShiftRepo
	client     *gomatrix.Client
}

func NewShiftCmd(repo model.ShiftRepo, client *gomatrix.Client) (*Shift, error) {
	regex := regexp.MustCompile(startShiftParseRegexString)

	return &Shift{
		parseRegex: regex,
		shiftRepo:  repo,
		client:     client,
	}, nil
}

func (s *Shift) Match(message string) bool {
	return s.parseRegex.MatchString(message)
}

func (s *Shift) Handle(event *gomatrix.Event) error {
	raw, ok := event.Content["body"].(string)
	if !ok {
		return matrix.ErrInvalidBody
	}

	parts := strings.Split(raw, " ")

	switch parts[0] {
	case "!startshift":
		return s.startShift(event, parts)
	case "!endshift":
		return s.endShift(event, parts)
	case "!listshifts":
		return s.listShifts(event)
	}

	return nil
}

func (s *Shift) startShift(event *gomatrix.Event, parts []string) error {
	if len(parts) < minCreateShiftLength {
		return matrix.ErrInvalidCommand
	}

	active, err := s.shiftRepo.Active(event.RoomID)
	if err != nil {
		return errors.Wrap(err, "error getting active shifts")
	}

	if len(active) > 0 {
		if _, err = s.client.SendText(event.RoomID, message.ActiveShiftOngoing); err != nil {
			return errors.Wrap(err, "error sending active shift message")
		}

		return nil
	}

	now := time.Now()

	if err = s.shiftRepo.Create(&model.Shift{
		RoomID:    event.RoomID,
		Sender:    event.Sender,
		Holders:   parts[1],
		StartTime: now,
		EndTime:   nil,
	}); err != nil {
		return errors.Wrap(err, "error saving shift")
	}

	_, err = s.client.SendFormattedText(event.RoomID, "",
		fmt.Sprintf(message.ShiftStarted, now.Local().Format(time.RFC850), message.FormatCommaSeperatedList(parts[1])))
	if err != nil {
		return errors.Wrap(err, "error sending shift created message")
	}

	return nil
}

func (s *Shift) endShift(event *gomatrix.Event, parts []string) error {
	if len(parts) < minEndShiftLength {
		return matrix.ErrInvalidCommand
	}

	shiftID, err := strconv.Atoi(parts[1])
	if err != nil {
		return errors.Wrap(err, "invalid shift id")
	}

	now := time.Now()

	if err = s.shiftRepo.Update(&model.Shift{
		ID:      shiftID,
		EndTime: &now,
	}); err != nil {
		return errors.Wrap(err, "error updating shift")
	}

	if _, err = s.client.SendFormattedText(event.RoomID, "", fmt.Sprintf(message.ShiftEnd, shiftID)); err != nil {
		return errors.Wrap(err, "error sending shift end message")
	}

	return nil
}

func (s *Shift) listShifts(event *gomatrix.Event) error {
	list, err := s.shiftRepo.Get(event.RoomID)
	if err != nil {
		return errors.Wrap(err, "error getting shifts")
	}

	msg := new(strings.Builder)

	for _, item := range list {
		end := "-"
		emoji := "🟢"

		if item.EndTime != nil {
			end = item.EndTime.Local().Format(time.RFC850)
			emoji = "⚪️"
		}

		msg.WriteString(fmt.Sprintf(message.ShiftItem, emoji,
			item.StartTime.Local().Format(time.RFC850), end, message.FormatCommaSeperatedList(item.Holders), item.ID))
	}

	response := fmt.Sprintf(message.ShiftList, msg.String())

	if _, err = s.client.SendFormattedText(event.RoomID, "", response); err != nil {
		return errors.Wrap(err, "error sending shifts list")
	}

	return nil
}
