package grpc

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ngocan-dev/mangahub/manga-backend/cmd/domain/favorite"
	"github.com/ngocan-dev/mangahub/manga-backend/cmd/domain/history"
	"github.com/ngocan-dev/mangahub/manga-backend/cmd/domain/manga"
	"github.com/ngocan-dev/mangahub/manga-backend/internal/chapter"
	pb "github.com/ngocan-dev/mangahub/manga-backend/proto/manga"
)

// Server implements the MangaService gRPC server
type Server struct {
	pb.UnimplementedMangaServiceServer
	db             *sql.DB
	mangaService   *manga.Service
	historyService *history.Service
}

// NewServer creates a new gRPC server instance
func NewServer(db *sql.DB) *Server {
	mangaService := manga.NewService(db)
	chapterRepo := chapter.NewRepository(db)
	chapterService := chapter.NewService(chapterRepo)
	mangaService.SetChapterService(chapterService)
	favoriteRepo := favorite.NewRepository(db)
	historyRepo := history.NewRepository(db)
	historyService := history.NewService(historyRepo, chapterService, favoriteRepo, mangaService)

	return &Server{
		db:             db,
		mangaService:   mangaService,
		historyService: historyService,
	}
}

// GetManga retrieves manga information by ID
// Main Success Scenario:
// 1. Client service calls GetManga gRPC method
// 2. gRPC server receives request with manga ID
// 3. Server queries database for manga information
// 4. Server constructs protobuf response message
// 5. Server returns manga data to client
func (s *Server) GetManga(ctx context.Context, req *pb.GetMangaRequest) (*pb.GetMangaResponse, error) {
	// Step 2: gRPC server receives request with manga ID
	if req.MangaId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid manga_id: must be greater than 0")
	}

	// Step 3: Server queries database for manga information
	mangaDetail, err := s.mangaService.GetDetails(ctx, req.MangaId, nil)
	if err != nil {
		if errors.Is(err, manga.ErrMangaNotFound) {
			return nil, status.Errorf(codes.NotFound, "manga not found: id=%d", req.MangaId)
		}
		log.Printf("Error fetching manga: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch manga: %v", err)
	}

	// Step 4: Server constructs protobuf response message
	pbManga := s.convertToProtoManga(mangaDetail)

	// Step 5: Server returns manga data to client
	return &pb.GetMangaResponse{
		Manga: pbManga,
	}, nil
}

// SearchManga searches for manga using various criteria
// Main Success Scenario:
// 1. Client calls SearchManga with search criteria
// 2. gRPC server processes search parameters
// 3. Server executes database query with filters
// 4. Server constructs response with result list
// 5. Server returns paginated results to client
func (s *Server) SearchManga(ctx context.Context, req *pb.SearchMangaRequest) (*pb.SearchMangaResponse, error) {
	// Step 2: Process search parameters
	searchReq := manga.SearchRequest{
		Query:  req.Query,
		Status: req.Status,
		Page:   int(req.Page),
		Limit:  int(req.Limit),
	}
	// Handle genre filter (convert single genre to slice if provided)
	if req.Genre != "" {
		searchReq.Genres = []string{req.Genre}
	}

	// Set defaults
	if searchReq.Page < 1 {
		searchReq.Page = 1
	}
	if searchReq.Limit < 1 {
		searchReq.Limit = 20
	}

	// Step 3: Execute database query with filters
	searchResp, err := s.mangaService.Search(ctx, searchReq)
	if err != nil {
		log.Printf("Error searching manga: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to search manga: %v", err)
	}

	// Step 4: Construct response with result list
	pbResults := make([]*pb.Manga, 0, len(searchResp.Results))
	for _, m := range searchResp.Results {
		pbManga := s.convertMangaToProto(&m)
		pbResults = append(pbResults, pbManga)
	}

	// Step 5: Return paginated results to client
	return &pb.SearchMangaResponse{
		Results: pbResults,
		Total:   int32(searchResp.Total),
		Page:    int32(searchResp.Page),
		Limit:   int32(searchResp.Limit),
		Pages:   int32(searchResp.Pages),
	}, nil
}

// convertMangaToProto converts domain Manga model to protobuf message
func (s *Server) convertMangaToProto(m *manga.Manga) *pb.Manga {
	// Get genres - Manga has Genre (singular)
	genres := make([]string, 0)
	if m.Genre != "" {
		genres = []string{m.Genre}
	}

	// Format timestamps
	updatedAt := ""
	if !m.DateUpdated.IsZero() {
		updatedAt = m.DateUpdated.Format(time.RFC3339)
	}

	return &pb.Manga{
		Id:           m.ID,
		Name:         m.Name,
		Title:        m.Title,
		Author:       m.Author,
		Description:  m.Description,
		CoverImage:   m.Image,
		Status:       m.Status,
		ChapterCount: 0, // Chapter count not included in search results
		Genres:       genres,
		CreatedAt:    "",
		UpdatedAt:    updatedAt,
	}
}

// convertToProtoManga converts domain manga model to protobuf message
func (s *Server) convertToProtoManga(m *manga.MangaDetail) *pb.Manga {
	// Get genres - MangaDetail embeds Manga which has Genre (singular)
	genres := make([]string, 0)
	if m.Genre != "" {
		genres = []string{m.Genre}
	}

	// Format timestamps
	createdAt := ""
	updatedAt := ""
	if !m.DateUpdated.IsZero() {
		updatedAt = m.DateUpdated.Format(time.RFC3339)
	}

	return &pb.Manga{
		Id:           m.ID,
		Name:         m.Name,
		Title:        m.Title,
		Author:       m.Author,
		Description:  m.Description,
		CoverImage:   m.Image, // Use Image field from Manga
		Status:       m.Status,
		ChapterCount: int32(m.ChapterCount),
		Genres:       genres,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}

// UpdateProgress updates user's reading progress for a manga
// Main Success Scenario:
// 1. Client calls UpdateProgress with user and manga data
// 2. gRPC server validates request parameters
// 3. Server updates user_progress table
// 4. Server triggers TCP broadcast for real-time sync
// 5. Server returns success confirmation
func (s *Server) UpdateProgress(ctx context.Context, req *pb.UpdateProgressRequest) (*pb.UpdateProgressResponse, error) {
	// Step 2: Validate request parameters
	if req.UserId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: must be greater than 0")
	}
	if req.MangaId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid manga_id: must be greater than 0")
	}
	if req.CurrentChapter < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid current_chapter: must be greater than 0")
	}

	// Convert to domain request
	updateReq := history.UpdateProgressRequest{
		CurrentChapter: int(req.CurrentChapter),
	}

	// Step 3: Update user_progress table
	// Step 4: Trigger TCP broadcast for real-time sync
	if s.historyService == nil {
		return nil, status.Error(codes.Unavailable, "history service not configured")
	}
	updateResp, err := s.historyService.UpdateProgress(ctx, req.UserId, req.MangaId, updateReq)
	if err != nil {
		if errors.Is(err, history.ErrMangaNotFound) {
			return nil, status.Errorf(codes.NotFound, "manga not found: id=%d", req.MangaId)
		}
		if errors.Is(err, history.ErrMangaNotInLibrary) {
			return nil, status.Errorf(codes.FailedPrecondition, "manga not in library. add it to your library first")
		}
		if errors.Is(err, history.ErrInvalidChapterNumber) {
			return nil, status.Errorf(codes.InvalidArgument, "invalid chapter number: %v", err)
		}
		log.Printf("Error updating progress: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update progress: %v", err)
	}

	// Step 5: Return success confirmation
	return &pb.UpdateProgressResponse{
		Success:        true,
		Message:        updateResp.Message,
		CurrentChapter: int32(updateResp.UserProgress.CurrentChapter),
		Broadcasted:    updateResp.Broadcasted,
	}, nil
}
