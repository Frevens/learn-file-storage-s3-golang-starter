package database

import (
	"testing"

	"github.com/google/uuid"
)

func TestUpdateVideoStoresThumbnailURL(t *testing.T) {
	db, err := NewClient(":memory:")
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	userID := uuid.New()
	video, err := db.CreateVideo(CreateVideoParams{
		Title:       "Test",
		Description: "Description",
		UserID:      userID,
	})
	if err != nil {
		t.Fatalf("CreateVideo returned error: %v", err)
	}

	thumbURL := "http://localhost:8080/api/thumbnails/" + video.ID.String()
	video.ThumbnailURL = &thumbURL

	if err := db.UpdateVideo(video); err != nil {
		t.Fatalf("UpdateVideo returned error: %v", err)
	}

	updated, err := db.GetVideo(video.ID)
	if err != nil {
		t.Fatalf("GetVideo returned error: %v", err)
	}
	if updated.ThumbnailURL == nil {
		t.Fatal("ThumbnailURL is nil after update")
	}
	if *updated.ThumbnailURL != thumbURL {
		t.Fatalf("expected thumbnail URL %q, got %q", thumbURL, *updated.ThumbnailURL)
	}
}
