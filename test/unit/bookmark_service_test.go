package unit

import (
	"context"
	"testing"

	"github.com/fallrising/goku-cli/internal/bookmarks"
	"github.com/fallrising/goku-cli/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of BookmarkRepository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetByID(ctx context.Context, id int64) (*models.Bookmark, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockRepository) Update(ctx context.Context, bookmark *models.Bookmark) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockRepository) Delete(ctx context.Context, id int64) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockRepository) List(ctx context.Context, limit, offset int) ([]*models.Bookmark, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.Bookmark, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockRepository) ListAllTags(ctx context.Context) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockRepository) CountByHostname(ctx context.Context) (map[string]int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockRepository) CountByTag(ctx context.Context) (map[string]int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockRepository) GetLatest(ctx context.Context, limit int) ([]*models.Bookmark, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockRepository) CountAccessibility(ctx context.Context) (map[string]int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockRepository) TopHostnames(ctx context.Context, limit int) ([]models.HostnameCount, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockRepository) ListUniqueHostnames(ctx context.Context) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockRepository) CountCreatedLastNDays(ctx context.Context, days int) (map[string]int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockRepository) Count(ctx context.Context) (int, error) {
	//TODO implement me
	panic("implement me")
}

// Implement only the methods you need for your tests
func (m *MockRepository) Create(ctx context.Context, bookmark *models.Bookmark) error {
	args := m.Called(ctx, bookmark)
	return args.Error(0)
}

func (m *MockRepository) GetByURL(ctx context.Context, url string) (*models.Bookmark, error) {
	args := m.Called(ctx, url)
	return args.Get(0).(*models.Bookmark), args.Error(1)
}

// Add other methods as needed for your tests

func TestCreateBookmark(t *testing.T) {
	mockRepo := new(MockRepository)
	service := bookmarks.NewBookmarkService(mockRepo)

	bookmark := &models.Bookmark{
		URL:   "https://example.com",
		Title: "Example",
	}

	// Set up expectations
	mockRepo.On("GetByURL", mock.Anything, bookmark.URL).Return((*models.Bookmark)(nil), nil)
	mockRepo.On("Create", mock.Anything, bookmark).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(1).(*models.Bookmark)
		arg.ID = 1 // Simulate ID assignment
	})

	err := service.CreateBookmark(context.Background(), bookmark)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), bookmark.ID)
	mockRepo.AssertExpectations(t)
}

// Add more tests for other BookmarkService methods
