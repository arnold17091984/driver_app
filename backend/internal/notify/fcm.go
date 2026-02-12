package notify

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	"github.com/kento/driver/backend/internal/model"
	"github.com/kento/driver/backend/internal/repository"
)

// FCMService sends push notifications via Firebase Cloud Messaging.
// If credPath is empty, all methods are no-ops (graceful degradation).
type FCMService struct {
	client   *messaging.Client
	userRepo *repository.UserRepo
}

func NewFCMService(credPath string, userRepo *repository.UserRepo) (*FCMService, error) {
	if credPath == "" {
		log.Println("[notify] Firebase credentials path not set, FCM disabled")
		return &FCMService{client: nil, userRepo: userRepo}, nil
	}

	opt := option.WithCredentialsFile(credPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, err
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, err
	}

	log.Println("[notify] FCM initialized successfully")
	return &FCMService{client: client, userRepo: userRepo}, nil
}

// NotifyUser sends a push notification to a specific user by user ID.
func (s *FCMService) NotifyUser(ctx context.Context, userID, title, body string, data map[string]string) {
	if s.client == nil {
		return
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil || user.FCMToken == nil {
		return
	}

	s.sendToToken(ctx, *user.FCMToken, title, body, data)
}

// NotifyVehicleDriver sends a push notification to the driver assigned to a vehicle.
func (s *FCMService) NotifyVehicleDriver(ctx context.Context, vehicleID, title, body string, data map[string]string) {
	if s.client == nil {
		return
	}

	drivers, err := s.userRepo.GetDriversByVehicleIDs(ctx, []string{vehicleID})
	if err != nil || len(drivers) == 0 {
		return
	}

	for _, d := range drivers {
		if d.FCMToken != nil {
			s.sendToToken(ctx, *d.FCMToken, title, body, data)
		}
	}
}

// NotifyRole sends a push notification to all active users with the given role(s).
func (s *FCMService) NotifyRole(ctx context.Context, title, body string, data map[string]string, roles ...model.Role) {
	if s.client == nil {
		return
	}

	users, err := s.userRepo.GetByRole(ctx, roles...)
	if err != nil {
		return
	}

	for _, u := range users {
		if u.FCMToken != nil {
			go s.sendToToken(ctx, *u.FCMToken, title, body, data)
		}
	}
}

func (s *FCMService) sendToToken(ctx context.Context, token, title, body string, data map[string]string) {
	msg := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Notification: &messaging.AndroidNotification{
				Sound: "default",
			},
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}

	if _, err := s.client.Send(ctx, msg); err != nil {
		log.Printf("[notify] FCM send error: %v", err)
	}
}
