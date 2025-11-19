-- Migration: Performance Indexes for Query Optimization
-- Goal: Optimize database queries for performance
-- Success Criteria: Indexes improve query performance

-- Indexes for Reading_Progress table
-- Used in: UpdateProgress, GetReadingStatistics, CalculateReadingStreaks
CREATE INDEX IF NOT EXISTS idx_reading_progress_user_novel ON Reading_Progress(User_Id, Novel_Id);
CREATE INDEX IF NOT EXISTS idx_reading_progress_last_read ON Reading_Progress(Last_Read_At);
CREATE INDEX IF NOT EXISTS idx_reading_progress_user_date ON Reading_Progress(User_Id, Last_Read_At);

-- Indexes for User_Library table
-- Used in: AddToLibrary, GetLibrary, GetFriendsActivities, IsMangaCompleted
CREATE INDEX IF NOT EXISTS idx_user_library_user_novel ON User_Library(User_Id, Novel_Id);
CREATE INDEX IF NOT EXISTS idx_user_library_status ON User_Library(User_Id, Status);
CREATE INDEX IF NOT EXISTS idx_user_library_completed ON User_Library(User_Id, Status, Completed_At) WHERE Status = 'completed';
CREATE INDEX IF NOT EXISTS idx_user_library_started ON User_Library(User_Id, Started_At) WHERE Started_At IS NOT NULL;

-- Indexes for Reviews table
-- Used in: GetReviews, GetReviewStats, GetFriendsActivities
CREATE INDEX IF NOT EXISTS idx_reviews_novel ON Reviews(Novel_Id, Created_At);
CREATE INDEX IF NOT EXISTS idx_reviews_user_novel ON Reviews(User_Id, Novel_Id);
CREATE INDEX IF NOT EXISTS idx_reviews_user_created ON Reviews(User_Id, Created_At);

-- Indexes for Novels table (for search)
-- Used in: Search, GetDetails
CREATE INDEX IF NOT EXISTS idx_novels_genre ON Novels(Genre) WHERE Genre IS NOT NULL AND Genre != '';
CREATE INDEX IF NOT EXISTS idx_novels_status ON Novels(Status) WHERE Status IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_novels_rating ON Novels(Rating) WHERE Rating IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_novels_title ON Novels(Title) WHERE Title IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_novels_author ON Novels(Author) WHERE Author IS NOT NULL;

-- Indexes for Friends table
-- Used in: GetFriends, GetFriendsActivities
CREATE INDEX IF NOT EXISTS idx_friends_user_status ON Friends(User_Id, Status);
CREATE INDEX IF NOT EXISTS idx_friends_friend_status ON Friends(Friend_Id, Status);
CREATE INDEX IF NOT EXISTS idx_friends_bidirectional ON Friends(User_Id, Friend_Id, Status);

-- Indexes for Progress_History table
-- Used in: CalculateReadingStatistics, CalculateReadingStreaks
CREATE INDEX IF NOT EXISTS idx_progress_history_user_date ON Progress_History(User_Id, Created_At);
CREATE INDEX IF NOT EXISTS idx_progress_history_novel ON Progress_History(Novel_Id, Created_At);

-- Indexes for Chapters table
-- Used in: GetDetails, ValidateChapterNumber
CREATE INDEX IF NOT EXISTS idx_chapters_novel_number ON Chapters(Novel_Id, Chapter_Number);
CREATE INDEX IF NOT EXISTS idx_chapters_novel ON Chapters(Novel_Id);

-- Indexes for Rating_System table
-- Used in: GetFriendsActivities
CREATE INDEX IF NOT EXISTS idx_rating_system_user_novel ON Rating_System(User_Id, Novel_Id);
CREATE INDEX IF NOT EXISTS idx_rating_system_user_date ON Rating_System(User_Id, Rating_Date);

-- Composite index for search queries (covers multiple search criteria)
-- Used in: Search function with multiple filters
CREATE INDEX IF NOT EXISTS idx_novels_search_composite ON Novels(Status, Genre, Rating) 
WHERE Status IS NOT NULL AND Genre IS NOT NULL AND Rating IS NOT NULL;

