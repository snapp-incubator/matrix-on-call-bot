package command

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/matrix-org/gomatrix"
	"github.com/pkg/errors"
	"github.com/snapp-incubator/matrix-on-call-bot/internal/matrix/message"
	"github.com/snapp-incubator/matrix-on-call-bot/internal/model"
)

const (
	// check if command starts with
	// !startshift or
	// !endshift or
	// !listshifts.
	startShiftParseRegexString = "^!startshift|^!endshift|^!listshifts"

	minCreateShiftLength int = 1

	minEndShiftLength int = 2
)

// Regexp is a compiled regular expression that can extract data in a message containing people mentioning (like:
// @ahmad.anvari:snapp.cab).
var mentionRegex = regexp.MustCompile(`<a href="https://matrix.to/#/(.*?)">(.*?)</a>`)

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

func (s *Shift) Handle(event *gomatrix.Event, messageParts []string) error {
	switch messageParts[0] {
	case "!startshift":
		return s.startShift(event, messageParts)
	case "!endshift":
		return s.endShift(event, messageParts)
	case "!listshifts":
		return s.listShifts(event)
	}

	return nil
}

//nolint:funlen,cyclop
func (s *Shift) startShift(event *gomatrix.Event, parts []string) error {
	if len(parts) < minCreateShiftLength {
		return ErrInvalidCommand
	}

	formattedBody, found := event.Content["formatted_body"]
	if !found && len(parts) > 1 {
		if _, err := s.client.SendText(event.RoomID, message.InvalidShiftStart); err != nil {
			return errors.Wrap(err, "error sending invalid shift start")
		}
	}

	mxids := make([]string, 0)
	names := make([]string, 0)
	mentions := make([]string, 0)

	if len(parts) == 1 {
		mxids = append(mxids, event.Sender)

		displayName, err := s.client.GetDisplayName(event.Sender)
		if err != nil {
			return errors.Wrap(err, "error getting the display name of the event sender")
		}

		mentions = append(mentions, message.MentionedText(event.Sender, displayName.DisplayName))
		names = append(names, displayName.DisplayName)
	} else {
		formattedBodyStr, ok := formattedBody.(string)
		if !ok {
			return errors.Wrap(ErrInvalidType, "error getting the display name of the event sender")
		}

		items := mentionRegex.FindAllStringSubmatch(formattedBodyStr, -1)
		for _, parts := range items {
			mentions = append(mentions, parts[0])
			mxids = append(mxids, parts[1])
			names = append(names, parts[2])
		}
	}

	active, err := s.shiftRepo.Active(event.RoomID)
	if err != nil {
		return errors.Wrap(err, "error getting active shifts")
	}

	//nolint:godox
	// TODO: Check weather this condition should be checked or not
	if len(active) > 0 {
		if _, err := s.client.SendText(event.RoomID, message.ActiveShiftOngoing); err != nil {
			return errors.Wrap(err, "error sending active shift message")
		}

		return nil
	}

	now := time.Now()

	for _, mxid := range mxids {
		//nolint:godox
		// TODO: Create a bulk Create method
		if err := s.shiftRepo.Create(&model.Shift{
			RoomID:    event.RoomID,
			Sender:    event.Sender,
			Holders:   mxid,
			StartTime: now,
			EndTime:   nil,
		}); err != nil {
			return errors.Wrap(err, "error saving shift")
		}
	}

	formattedTime := now.Local().Format(time.RFC850)

	_, err = s.client.SendFormattedText(event.RoomID, fmt.Sprintf(message.ShiftStarted,
		strings.Join(names, " "), formattedTime),
		fmt.Sprintf(message.ShiftStarted, strings.Join(mentions, " "), formattedTime))
	if err != nil {
		return errors.Wrap(err, "error sending shift created message")
	}

	return nil
}

func (s *Shift) endShift(event *gomatrix.Event, parts []string) error {
	if len(parts) < minEndShiftLength {
		return ErrInvalidCommand
	}

	shiftID, err := strconv.Atoi(parts[1])
	if err != nil {
		return errors.Wrap(err, "invalid shift id")
	}

	now := time.Now()

	if err := s.shiftRepo.Update(&model.Shift{
		ID:      shiftID,
		EndTime: &now,
	}); err != nil {
		return errors.Wrap(err, "error updating shift")
	}

	_, err = s.client.SendFormattedText(event.RoomID,
		fmt.Sprintf(message.ShiftEnd, shiftID), fmt.Sprintf(message.ShiftEndFormatted, shiftID))
	if err != nil {
		return errors.Wrap(err, "error sending shift end message")
	}

	return nil
}

func (s *Shift) listShifts(event *gomatrix.Event) error {
	list, err := s.shiftRepo.Get(event.RoomID)
	if err != nil {
		return errors.Wrap(err, "error getting shifts")
	}

	msg := ""

	for _, item := range list {
		end := "-"
		emoji := "🟢"

		if item.EndTime != nil {
			end = item.EndTime.Local().Format(time.RFC850)
			emoji = "⚪️"
		}

		displayName, err := s.client.GetDisplayName(item.Holders)
		if err != nil {
			return errors.Wrap(err, "error getting the display name of the event sender")
		}

		msg += fmt.Sprintf(message.ShiftItem, emoji,
			item.StartTime.Local().Format(time.RFC850),
			end, message.MentionedText(item.Holders, displayName.DisplayName), item.ID)
	}

	msg = fmt.Sprintf(message.ShiftList, msg)

	if _, err = s.client.SendFormattedText(event.RoomID, "", msg); err != nil {
		return errors.Wrap(err, "error sending shifts list")
	}

	return nil
}
