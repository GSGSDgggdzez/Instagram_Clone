package utils

import (
	"os"
	"time"

	"github.com/posthog/posthog-go"
)

var client posthog.Client

// InitAnalytics - Fire up our analytics engine! 🚀
// InitAnalytics needs to be called when your app starts
func InitAnalytics() error {
	var err error
	client, err = posthog.NewWithConfig(
		os.Getenv("PostHog_API_KEY"), // Make sure this env var exists!
		posthog.Config{
			Endpoint: os.Getenv("PostHog_Url"), // And this one too!
		},
	)
	return err
}

// TrackRegistration - Monitor those sign-ups! 💪
func TrackRegistration(email string, success bool, retryCount int) {
	if client == nil {
		return // Safely handle nil client
	}
	client.Enqueue(posthog.Capture{
		DistinctId: email,
		Event:      "user_registration",
		Properties: posthog.NewProperties().
			Set("success", success).
			Set("retry_count", retryCount).
			Set("timestamp", time.Now()),
	})
}

func TrackLogin(email string, success bool, retryCount int) {
	if client == nil {
		return // Safely handle nil client
	}
	client.Enqueue(posthog.Capture{
		DistinctId: email,
		Event:      "user_Login",
		Properties: posthog.NewProperties().
			Set("success", success).
			Set("retry_count", retryCount).
			Set("timestamp", time.Now()),
	})
}

func TrackFindUserByID(email string, success bool, retryCount int) {
	if client == nil {
		return // Safely handle nil client
	}
	client.Enqueue(posthog.Capture{
		DistinctId: email,
		Event:      "FindUserByID",
		Properties: posthog.NewProperties().
			Set("success", success).
			Set("retry_count", retryCount).
			Set("timestamp", time.Now()),
	})
}

func TrackFindUserByEMAIL(email string, success bool, retryCount int) {
	if client == nil {
		return // Safely handle nil client
	}
	client.Enqueue(posthog.Capture{
		DistinctId: email,
		Event:      "FindUserByEmail",
		Properties: posthog.NewProperties().
			Set("success", success).
			Set("retry_count", retryCount).
			Set("timestamp", time.Now()),
	})
}

func TrackFindUserByToken(email string, success bool, retryCount int) {
	if client == nil {
		return // Safely handle nil client
	}
	client.Enqueue(posthog.Capture{
		DistinctId: email,
		Event:      "FindUserByToken",
		Properties: posthog.NewProperties().
			Set("success", success).
			Set("retry_count", retryCount).
			Set("timestamp", time.Now()),
	})
}

func TrackUserDeletion(email string, success bool, retryCount int) {
	if client == nil {
		return // Safely handle nil client
	}
	client.Enqueue(posthog.Capture{
		DistinctId: email,
		Event:      "DeleteUser",
		Properties: posthog.NewProperties().
			Set("success", success).
			Set("retry_count", retryCount).
			Set("timestamp", time.Now()),
	})
}

func TrackRelationshipDeletion(email string, success bool, retryCount int) {
	if client == nil {
		return // Safely handle nil client
	}
	client.Enqueue(posthog.Capture{
		DistinctId: email,
		Event:      "DeleteUserRelationships",
		Properties: posthog.NewProperties().
			Set("success", success).
			Set("retry_count", retryCount).
			Set("timestamp", time.Now()),
	})
}

func TrackPostDeletion(email string, success bool, retryCount int) {
	if client == nil {
		return // Safely handle nil client
	}
	client.Enqueue(posthog.Capture{
		DistinctId: email,
		Event:      "DeleteUserPosts",
		Properties: posthog.NewProperties().
			Set("success", success).
			Set("retry_count", retryCount).
			Set("timestamp", time.Now()),
	})
}

func TrackCommentDeletion(email string, success bool, retryCount int) {
	if client == nil {
		return // Safely handle nil client
	}
	client.Enqueue(posthog.Capture{
		DistinctId: email,
		Event:      "DeleteUserComments",
		Properties: posthog.NewProperties().
			Set("success", success).
			Set("retry_count", retryCount).
			Set("timestamp", time.Now()),
	})
}

// Add more tracking functions as needed! 🎯
