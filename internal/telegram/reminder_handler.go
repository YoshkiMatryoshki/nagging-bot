package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"naggingbot/internal/domain"
)

// ReminderHandler handles /reminder command to create a reminder for a user.
// Format: /reminder Name_Description_StartDate_EndDate_HH:MM;HH:MM_TimeZone
type ReminderHandler struct {
	users       domain.UserStore
	reminders   domain.ReminderStore
	occurrences domain.OccurrenceStore
	responder   Responder
}

func NewReminderHandler(users domain.UserStore, reminders domain.ReminderStore, occurrences domain.OccurrenceStore, responder Responder) *ReminderHandler {
	return &ReminderHandler{
		users:       users,
		reminders:   reminders,
		occurrences: occurrences,
		responder:   responder,
	}
}

func (h *ReminderHandler) HandleCommand(ctx context.Context, msg *Message) error {
	user := msg.From
	if user == nil {
		return nil
	}
	// Format: /reminder Name_Description_StartDate_EndDate_HH:MM;HH:MM_TimeZone
	parts := strings.SplitN(strings.TrimSpace(msg.Text), " ", 2)
	if len(parts) < 2 {
		h.reply(ctx, user.ID, "Usage: /reminder Name_Description_StartDate_EndDate_HH:MM;HH:MM_TimeZone\nExample: /reminder Pill_VitC_19.01.2026_20.01.2026_08:00;13:00;19:00_Europe/Warsaw")
		return nil
	}
	payload := parts[1]
	fields := strings.SplitN(payload, "_", 6)
	if len(fields) != 6 {
		h.reply(ctx, user.ID, "Invalid format. Expected: /reminder Name_Description_StartDate_EndDate_HH:MM;HH:MM_TimeZone")
		return nil
	}

	name := fields[0]
	description := fields[1]
	startDateStr := fields[2]
	endDateStr := fields[3]
	timesStr := fields[4]
	timezone := fields[5]

	tod, err := parseTimesOfDay(timesStr)
	if err != nil {
		h.reply(ctx, user.ID, "Invalid times. Use HH:MM;HH:MM")
		return nil
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		h.reply(ctx, user.ID, "Invalid timezone. Use IANA, e.g., Europe/Moscow")
		return nil
	}

	// Ensure user exists.
	domainUser := &domain.User{
		TelegramID: user.ID,
		Username:   user.Username,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Language:   user.LanguageCode,
	}
	if err := h.users.Upsert(ctx, domainUser); err != nil {
		log.Printf("telegram: reminder upsert user failed: %v", err)
		h.reply(ctx, user.ID, "Failed to save user")
		return nil
	}

	start, end, err := parseDateRange(startDateStr, endDateStr, timezone)
	if err != nil {
		h.reply(ctx, user.ID, "Invalid date range. Use DD.MM.YYYY_DD.MM.YYYY (inclusive)")
		return nil
	}

	rem := &domain.Reminder{
		UserID:      domainUser.ID,
		Name:        name,
		Description: description,
		StartDate:   start,
		EndDate:     end,
		TimesOfDay:  tod,
		TimeZone:    timezone,
		IsActive:    true,
	}
	if err := h.reminders.Create(ctx, rem); err != nil {
		log.Printf("telegram: create reminder failed: %v", err)
		h.reply(ctx, user.ID, "Failed to create reminder")
		return nil
	}

	// Create occurrences for the date range based on times of day and timezone.
	if err := h.seedOccurrences(ctx, rem, loc); err != nil {
		log.Printf("telegram: create occurrences failed: %v", err)
		h.reply(ctx, user.ID, "Reminder created, but failed to schedule occurrences")
		return nil
	}

	h.reply(ctx, user.ID, fmt.Sprintf("Reminder created: %s (%s) in %s", name, description, timezone))
	return nil
}

func (h *ReminderHandler) reply(ctx context.Context, chatID int64, text string) {
	if h.responder == nil {
		return
	}
	if err := h.responder.SendMessage(ctx, chatID, text); err != nil {
		log.Printf("telegram: failed to send reply: %v", err)
	}
}

func parseTimesOfDay(s string) ([]domain.TimeOfDay, error) {
	if strings.TrimSpace(s) == "" {
		return nil, fmt.Errorf("empty times")
	}
	parts := strings.Split(s, ";")
	var out []domain.TimeOfDay
	for _, p := range parts {
		p = strings.TrimSpace(p)
		t, err := time.Parse("15:04", p)
		if err != nil {
			return nil, err
		}
		out = append(out, domain.TimeOfDay{Hour: t.Hour(), Minute: t.Minute()})
	}
	return out, nil
}

func (h *ReminderHandler) seedOccurrences(ctx context.Context, rem *domain.Reminder, loc *time.Location) error {
	// iterate each day in local tz from start to end inclusive
	startLoc := rem.StartDate.In(loc)
	endLoc := rem.EndDate.In(loc)

	for day := startLoc; !day.After(endLoc); day = day.Add(24 * time.Hour) {
		for _, tod := range rem.TimesOfDay {
			fire := time.Date(day.Year(), day.Month(), day.Day(), tod.Hour, tod.Minute, 0, 0, loc)
			occ := &domain.Occurrence{
				ReminderID: rem.ID,
				FireAtUtc:  fire.UTC(),
				Status:     domain.OccurrenceCreated,
			}
			if err := h.occurrences.Create(ctx, occ); err != nil {
				log.Printf("telegram: failed to create occurrence at %s: %v", fire.UTC(), err)
			}
		}
	}
	return nil
}

func parseDateRange(startStr, endStr, tz string) (time.Time, time.Time, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	start, err := time.ParseInLocation("02.01.2006", strings.TrimSpace(startStr), loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := time.ParseInLocation("02.01.2006", strings.TrimSpace(endStr), loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("end before start")
	}
	// Inclusive end-of-day in local TZ.
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, loc)
	return start.UTC(), end.UTC(), nil
}
