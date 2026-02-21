package calendar

import (
	"bincang-visual/internal/domain/entity"
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type GoogleCalendarRepository struct {
	config *oauth2.Config
}

func NewGoogleCalendarRepository(config *oauth2.Config) *GoogleCalendarRepository {
	return &GoogleCalendarRepository{
		config: config,
	}
}

func (r *GoogleCalendarRepository) CreateEvent(ctx context.Context, event *entity.CalendarEvent) error {
	// Get OAuth2 token from context (set by middleware)
	token, ok := ctx.Value("oauth_token").(*oauth2.Token)
	if !ok {
		return fmt.Errorf("no oauth token in context")
	}

	client := r.config.Client(ctx, token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to create calendar service: %w", err)
	}

	calendarEvent := &calendar.Event{
		Summary:     event.Title,
		Description: event.Description + "\n\nMeeting URL: " + r.getMeetingURL(event.RoomID),
		Start: &calendar.EventDateTime{
			DateTime: event.StartTime.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		End: &calendar.EventDateTime{
			DateTime: event.EndTime.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		ConferenceData: &calendar.ConferenceData{
			CreateRequest: &calendar.CreateConferenceRequest{
				RequestId: event.RoomID,
				ConferenceSolutionKey: &calendar.ConferenceSolutionKey{
					Type: "hangoutsMeet",
				},
			},
		},
		Attendees: r.convertAttendees(event.Attendees),
	}

	createdEvent, err := srv.Events.Insert("primary", calendarEvent).
		ConferenceDataVersion(1).
		Do()
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	event.GoogleEventID = createdEvent.Id
	return nil
}

func (r *GoogleCalendarRepository) GetEvent(ctx context.Context, eventID string) (*entity.CalendarEvent, error) {
	token, ok := ctx.Value("oauth_token").(*oauth2.Token)
	if !ok {
		return nil, fmt.Errorf("no oauth token in context")
	}

	client := r.config.Client(ctx, token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	event, err := srv.Events.Get("primary", eventID).Do()
	if err != nil {
		return nil, err
	}

	return r.convertFromGoogleEvent(event), nil
}

func (r *GoogleCalendarRepository) UpdateEvent(ctx context.Context, event *entity.CalendarEvent) error {
	token, ok := ctx.Value("oauth_token").(*oauth2.Token)
	if !ok {
		return fmt.Errorf("no oauth token in context")
	}

	client := r.config.Client(ctx, token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return err
	}

	calendarEvent := &calendar.Event{
		Summary:     event.Title,
		Description: event.Description,
		Start: &calendar.EventDateTime{
			DateTime: event.StartTime.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: event.EndTime.Format(time.RFC3339),
		},
	}

	_, err = srv.Events.Update("primary", event.GoogleEventID, calendarEvent).Do()
	return err
}

func (r *GoogleCalendarRepository) DeleteEvent(ctx context.Context, eventID string) error {
	token, ok := ctx.Value("oauth_token").(*oauth2.Token)
	if !ok {
		return fmt.Errorf("no oauth token in context")
	}

	client := r.config.Client(ctx, token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return err
	}

	return srv.Events.Delete("primary", eventID).Do()
}

func (r *GoogleCalendarRepository) GetUserEvents(ctx context.Context, userID string, from, to time.Time) ([]*entity.CalendarEvent, error) {
	token, ok := ctx.Value("oauth_token").(*oauth2.Token)
	if !ok {
		return nil, fmt.Errorf("no oauth token in context")
	}

	client := r.config.Client(ctx, token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	events, err := srv.Events.List("primary").
		TimeMin(from.Format(time.RFC3339)).
		TimeMax(to.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, err
	}

	result := make([]*entity.CalendarEvent, 0, len(events.Items))
	for _, item := range events.Items {
		result = append(result, r.convertFromGoogleEvent(item))
	}

	return result, nil
}

func (r *GoogleCalendarRepository) convertAttendees(emails []string) []*calendar.EventAttendee {
	attendees := make([]*calendar.EventAttendee, len(emails))
	for i, email := range emails {
		attendees[i] = &calendar.EventAttendee{
			Email: email,
		}
	}
	return attendees
}

func (r *GoogleCalendarRepository) convertFromGoogleEvent(event *calendar.Event) *entity.CalendarEvent {
	startTime, _ := time.Parse(time.RFC3339, event.Start.DateTime)
	endTime, _ := time.Parse(time.RFC3339, event.End.DateTime)

	attendees := make([]string, len(event.Attendees))
	for i, a := range event.Attendees {
		attendees[i] = a.Email
	}

	return &entity.CalendarEvent{
		ID:            event.Id,
		GoogleEventID: event.Id,
		Title:         event.Summary,
		Description:   event.Description,
		StartTime:     startTime,
		EndTime:       endTime,
		Attendees:     attendees,
	}
}

func (r *GoogleCalendarRepository) getMeetingURL(roomID string) string {
	// TODO: replace base url
	return fmt.Sprintf("https://my-domain.com/join/%s", roomID)
}
