package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	meetingIDLength = 8
	passwordLength  = 6
	meetingTTL      = 24 * time.Hour
	meetingCollName = "meetings"
)

// Meeting represents a stored meeting session.
type Meeting struct {
	ID        string               `bson:"_id" json:"id"`
	Password  string               `bson:"password" json:"password"`
	CreatedAt time.Time            `bson:"createdAt" json:"createdAt"`
	ExpiresAt time.Time            `bson:"expiresAt" json:"expiresAt"`
	Metadata  map[string]any       `bson:"metadata,omitempty" json:"metadata,omitempty"`
	Members   []primitive.ObjectID `bson:"members" json:"members"`
}

// MeetingStore manages meeting persistence in MongoDB.
type MeetingStore struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// NewMeetingStore creates a new MeetingStore.
func NewMeetingStore(ctx context.Context, uri, dbName string) (*MeetingStore, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	coll := client.Database(dbName).Collection(meetingCollName)

	// Ensure TTL index
	_, _ = coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expiresAt", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})

	return &MeetingStore{client: client, collection: coll}, nil
}

// Disconnect closes the MongoDB client.
func (s *MeetingStore) Disconnect(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}

// CreateMeeting generates a new meeting and persists it.
func (s *MeetingStore) CreateMeeting(ctx context.Context) (*Meeting, error) {
	meetingID, err := randomAlphaNumeric(meetingIDLength)
	if err != nil {
		return nil, err
	}
	password, err := randomDigits(passwordLength)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	meeting := &Meeting{
		ID:        meetingID,
		Password:  password,
		CreatedAt: now,
		ExpiresAt: now.Add(meetingTTL),
		Members:   []primitive.ObjectID{},
	}

	_, err = s.collection.InsertOne(ctx, meeting)
	if mongo.IsDuplicateKeyError(err) {
		// Collision, retry recursively.
		return s.CreateMeeting(ctx)
	}
	return meeting, err
}

// ValidateMeeting checks if the meeting exists and password matches.
func (s *MeetingStore) ValidateMeeting(ctx context.Context, meetingID, password string) (*Meeting, error) {
	var meeting Meeting
	err := s.collection.FindOne(ctx, bson.M{"_id": meetingID, "password": password}).Decode(&meeting)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrMeetingNotFound
		}
		return nil, err
	}
	return &meeting, nil
}

// ErrMeetingNotFound indicates the meeting cannot be found or invalid password.
var ErrMeetingNotFound = errors.New("meeting not found or invalid password")

func randomAlphaNumeric(length int) (string, error) {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	return randomStringFromCharset(length, chars)
}

func randomDigits(length int) (string, error) {
	const digits = "0123456789"
	return randomStringFromCharset(length, digits)
}

func randomStringFromCharset(length int, charset string) (string, error) {
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}

// MustRandomID is exported for tests or fallback operations.
func MustRandomID(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)[:length]
}
