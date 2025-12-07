package monitor

import (
	"log/slog"
	"time"
)

type Monitor struct {
	cCalendarReminder chan CalendarReminder
	cEmail            chan Email

	stop chan struct{}
}

type CalendarReminder struct {
}

type Email struct {
}

func NewMonitor() *Monitor {
	return &Monitor{
		cCalendarReminder: make(chan CalendarReminder),
		cEmail:            make(chan Email),
		stop:              make(chan struct{}),
	}
}

func (m *Monitor) CalendarReminder() <-chan CalendarReminder {
	return m.cCalendarReminder
}

func (m *Monitor) Email() <-chan Email {
	return m.cEmail
}

func (m *Monitor) Run() {
	t := time.NewTicker(time.Second)
	for {
		select {
		case <-t.C:
			pollCalendarReminders()
			pollEmails()
		case <-m.stop:
			close(m.cCalendarReminder)
			close(m.cEmail)
			return
		}
	}
}

func (m *Monitor) Stop() {
	close(m.stop)
}

func pollCalendarReminders() {
	slog.Debug("polling calendar reminders")
	// TODO
}

func pollEmails() {
	slog.Debug("polling emails")
	// TODO
}
