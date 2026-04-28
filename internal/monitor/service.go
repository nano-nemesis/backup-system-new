package monitor

import (
	"context"
	"fmt"
	"time"

	"backup-system/config"
	"backup-system/internal/notifier"
	"backup-system/internal/utils"
)

type Service struct {
	cfg      config.Config
	logger   *utils.Logger
	notifier *notifier.Telegram
	state    *StateStore
	watcher  *Watcher
}

func NewService(cfg config.Config, logger *utils.Logger) *Service {
	return &Service{
		cfg:      cfg,
		logger:   logger,
		notifier: notifier.NewTelegram(cfg.TelegramBotToken, cfg.TelegramChatID),
		state:    NewStateStore(cfg.AlertCooldown),
		watcher:  NewWatcher(cfg.LogDir, cfg.MonitorPollInterval, logger),
	}
}

func (s *Service) Run(ctx context.Context) error {
	s.logger.Printf("monitor started on %s", s.cfg.LogDir)
	return s.watcher.Watch(ctx, s.handleEvent)
}

func (s *Service) handleEvent(event Event) {
	now := time.Now()

	switch event.Status {
	case StatusError:
		s.state.MarkError(event.NodeName)
		if !s.state.ShouldAlert(event.NodeName, now) {
			return
		}

		message := fmt.Sprintf(
			"BACKUP ALERT\nnode: %s\ncategory: %s\naction: %s\nlog: %s",
			event.NodeName,
			event.Category,
			event.Action,
			event.Line,
		)
		if err := s.notifier.Send(message); err != nil {
			s.logger.Printf("monitor telegram send failed for %s: %v", event.NodeName, err)
			return
		}
		s.logger.Printf("monitor alert sent for %s", event.NodeName)
	case StatusSuccess:
		if !s.state.MarkSuccess(event.NodeName) {
			return
		}

		message := fmt.Sprintf(
			"BACKUP RECOVERY\nnode: %s\nstatus: normal\nlog: %s",
			event.NodeName,
			event.Line,
		)
		if err := s.notifier.Send(message); err != nil {
			s.logger.Printf("monitor telegram recovery send failed for %s: %v", event.NodeName, err)
			return
		}
		s.logger.Printf("monitor recovery sent for %s", event.NodeName)
	}
}
