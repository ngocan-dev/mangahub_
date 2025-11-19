# Domain Refactoring Plan

## Current Problem
The `manga` domain is doing too much - it contains:
- Manga CRUD operations ✅ (should stay)
- Reviews/Reviews management ❌ (should move to `comment` or `review`)
- Ratings ❌ (should move to `rating`)
- Reading progress/history ❌ (should move to `history`)
- Favorites ❌ (should move to `favorite`)
- Chapter operations ❌ (should move to `chapter`)

## Proposed Domain Structure

### 1. `manga` Domain (Core Manga Entity)
**Responsibility**: Manga metadata and search
- `Search()` - Search manga
- `GetDetails()` - Get manga details
- `GetByID()` - Get manga by ID
- Manga models (Manga, MangaDetail)

**Should NOT contain**:
- User-specific data (progress, favorites, ratings)
- Reviews
- Chapters

### 2. `rating` Domain
**Responsibility**: Rating management
- `CreateRating()` - Rate a manga
- `UpdateRating()` - Update rating
- `GetRating()` - Get user's rating
- `GetAverageRating()` - Get average rating for manga
- Rating models

**Move from manga**:
- `LibraryStatus.Rating` field
- Rating-related logic in reviews

### 3. `comment` Domain (or `review`)
**Responsibility**: Review/Comment management
- `CreateReview()` - Create review
- `GetReviews()` - Get reviews for manga
- `UpdateReview()` - Update review
- `DeleteReview()` - Delete review
- `GetReviewStats()` - Get review statistics
- Review models

**Move from manga**:
- `CreateReview()`, `GetReviews()`
- `Review`, `ReviewStats` models
- All review-related repository methods

### 4. `history` Domain
**Responsibility**: Reading progress and history
- `UpdateProgress()` - Update reading progress
- `GetProgress()` - Get user progress
- `GetReadingHistory()` - Get reading history
- `GetReadingStatistics()` - Calculate statistics
- `GetReadingAnalytics()` - Get analytics
- Progress models

**Move from manga**:
- `UpdateProgress()`
- `GetReadingStatistics()`, `GetReadingAnalytics()`
- `UserProgress` model
- Progress-related repository methods

### 5. `favorite` Domain
**Responsibility**: Favorite management
- `AddFavorite()` - Add to favorites
- `RemoveFavorite()` - Remove from favorites
- `GetFavorites()` - Get user favorites
- `IsFavorite()` - Check if favorite

**Move from manga**:
- `LibraryStatus.IsFavorite` field
- Favorite logic in `AddToLibrary()`

### 6. `chapter` Domain
**Responsibility**: Chapter management
- `GetChapters()` - Get chapters for manga
- `GetChapter()` - Get chapter by ID/number
- `ValidateChapter()` - Validate chapter exists
- `GetChapterCount()` - Get chapter count
- Chapter models

**Move from manga**:
- Chapter validation logic
- Chapter count operations
- Chapter-related repository methods

### 7. `library` Domain (Optional - could stay in manga)
**Responsibility**: User library management
- `AddToLibrary()` - Add manga to library
- `RemoveFromLibrary()` - Remove from library
- `GetLibrary()` - Get user library
- `UpdateLibraryStatus()` - Update status
- Library models

**Move from manga**:
- `AddToLibrary()`
- `LibraryStatus` model (except Rating and IsFavorite)

## Migration Strategy

### Phase 1: Create Domain Structures
1. Create models for each domain
2. Create repository interfaces
3. Create service interfaces

### Phase 2: Implement New Domains
1. Implement `rating` domain
2. Implement `comment` domain
3. Implement `history` domain
4. Implement `favorite` domain
5. Implement `chapter` domain

### Phase 3: Refactor Manga Domain
1. Remove moved functionality from manga
2. Update manga to use other domains via dependency injection
3. Update handlers to use new domains

### Phase 4: Update Handlers
1. Update HTTP handlers to use new domains
2. Update gRPC handlers
3. Update WebSocket handlers

## Benefits

1. **Single Responsibility**: Each domain has one clear purpose
2. **Maintainability**: Easier to find and modify code
3. **Testability**: Can test each domain independently
4. **Scalability**: Can scale domains independently
5. **Reusability**: Domains can be reused in different contexts

## Example: After Refactoring

```go
// manga/service.go - Only manga operations
type Service struct {
    repo *Repository
    cache MangaCacher
}

func (s *Service) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error)
func (s *Service) GetDetails(ctx context.Context, mangaID int64) (*MangaDetail, error)

// rating/service.go - Rating operations
type Service struct {
    repo *Repository
    mangaService manga.Service // Dependency
}

func (s *Service) CreateRating(ctx context.Context, userID, mangaID int64, rating int) error
func (s *Service) GetAverageRating(ctx context.Context, mangaID int64) (float64, error)

// comment/service.go - Review operations
type Service struct {
    repo *Repository
    mangaService manga.Service
    historyService history.Service // To check if completed
}

func (s *Service) CreateReview(ctx context.Context, userID, mangaID int64, req CreateReviewRequest) (*CreateReviewResponse, error)
func (s *Service) GetReviews(ctx context.Context, mangaID int64, page, limit int) (*GetReviewsResponse, error)

// history/service.go - Progress operations
type Service struct {
    repo *Repository
    chapterService chapter.Service
    broadcaster Broadcaster
}

func (s *Service) UpdateProgress(ctx context.Context, userID, mangaID int64, req UpdateProgressRequest) (*UpdateProgressResponse, error)
func (s *Service) GetReadingStatistics(ctx context.Context, userID int64) (*ReadingStatistics, error)
```

