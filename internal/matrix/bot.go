package matrix

import (
	"fmt"
	"strings"
	"time"

	"github.com/matrix-org/gomatrix"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/snapp-incubator/matrix-on-call-bot/internal/model"
)

const (
	RoomMemberEvent  = "m.room.member"
	RoomMessageEvent = "m.room.message"
	ResyncWaitTime   = 300 * time.Millisecond
)

var ErrBotSyncCreationFailed = errors.New("cannot create sync for bot")

type Bot struct {
	cli *gomatrix.Client

	displayName string
	userID      string
	autoJoin    bool

	roomRepo     model.RoomRepo
	shiftRepo    model.ShiftRepo
	followUpRepo model.FollowUpRepo

	commands []Command

	stopSignal chan struct{}
}

func New(url, userID, token, displayName string,
	roomRepo model.RoomRepo, shiftRepo model.ShiftRepo, followUpRepo model.FollowUpRepo,
) (*Bot, error) {
	cli, err := gomatrix.NewClient(url, userID, token)
	if err != nil {
		return nil, errors.Wrap(err, "can't create client")
	}

	return &Bot{
		cli:          cli,
		displayName:  displayName,
		userID:       userID,
		autoJoin:     true,
		roomRepo:     roomRepo,
		shiftRepo:    shiftRepo,
		followUpRepo: followUpRepo,
		stopSignal:   make(chan struct{}, 1),
	}, nil
}

func (b *Bot) RegisterListeners() error {
	syncer, ok := b.cli.Syncer.(*gomatrix.DefaultSyncer)
	if !ok {
		return ErrBotSyncCreationFailed
	}

	syncer.OnEventType(RoomMemberEvent, func(event *gomatrix.Event) {
		logrus.WithField("event", RoomMemberEvent).Debugf("got event: %+v", event)

		if !b.autoJoin {
			return
		}

		//nolint:nestif
		if val, ok := event.Content["membership"]; ok {
			if membership, ok := val.(string); ok && membership == "invite" {
				if err := b.roomRepo.Create(&model.Room{
					ID:     event.RoomID,
					Sender: event.Sender,
				}); err != nil {
					logrus.WithField("error", err.Error()).Error("error saving room")

					return
				}
				if _, err := b.cli.JoinRoom(event.RoomID, "", ""); err != nil {
					logrus.WithField("error", err.Error()).Error("error joining room")
				} else {
					logrus.WithField("room_id", event.RoomID).Info("joined room")
				}
			}
		}
	})

	syncer.OnEventType(RoomMessageEvent, func(event *gomatrix.Event) {
		if event.Sender == b.userID {
			return
		}

		if err := b.matchAndHandleEvent(event); err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err.Error(),
				"event": event,
			}).Error("error handling message")
		}
	})

	return nil
}

func (b *Bot) matchAndHandleEvent(event *gomatrix.Event) error {
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

	for _, cmd := range b.commands {
		if cmd.Match(parts[0]) {
			// each command message can be handled by only one command handler.
			err := cmd.Handle(event)
			if err != nil {
				return fmt.Errorf("command type %t returned error: %w", cmd, err)
			}

			return nil
		}
	}

	return ErrUnknownCommand
}

func (b *Bot) Run() {
	ticker := time.NewTicker(ResyncWaitTime)

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := b.cli.Sync(); err != nil {
					logrus.WithField("error", err.Error()).Error("sync failed")
				}
			case <-b.stopSignal:
				ticker.Stop()

				return
			}
		}
	}()
}

func (b *Bot) Stop() {
	b.cli.StopSync()
	b.stopSignal <- struct{}{}
}
