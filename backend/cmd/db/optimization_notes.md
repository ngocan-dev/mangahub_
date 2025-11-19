# Database Query Optimization Notes

## Indexing Recommendations

To ensure **database queries complete efficiently** and **API response times remain under 500ms**, the following indexes should be created:

### Critical Indexes for Concurrent Access (50-100 users)

1. **Reading_Progress table:**
   ```sql
   CREATE INDEX IF NOT EXISTS idx_reading_progress_user_novel ON Reading_Progress(User_Id, Novel_Id);
   CREATE INDEX IF NOT EXISTS idx_reading_progress_last_read ON Reading_Progress(Last_Read_At);
   ```

2. **User_Library table:**
   ```sql
   CREATE INDEX IF NOT EXISTS idx_user_library_user_novel ON User_Library(User_Id, Novel_Id);
   CREATE INDEX IF NOT EXISTS idx_user_library_status ON User_Library(User_Id, Status);
   CREATE INDEX IF NOT EXISTS idx_user_library_completed ON User_Library(User_Id, Status, Completed_At) WHERE Status = 'completed';
   ```

3. **Reviews table:**
   ```sql
   CREATE INDEX IF NOT EXISTS idx_reviews_novel ON Reviews(Novel_Id, Created_At);
   CREATE INDEX IF NOT EXISTS idx_reviews_user_novel ON Reviews(User_Id, Novel_Id);
   ```

4. **Novels table (for search):**
   ```sql
   CREATE INDEX IF NOT EXISTS idx_novels_genre ON Novels(Genre);
   CREATE INDEX IF NOT EXISTS idx_novels_status ON Novels(Status);
   CREATE INDEX IF NOT EXISTS idx_novels_rating ON Novels(Rating);
   ```

5. **Friends table:**
   ```sql
   CREATE INDEX IF NOT EXISTS idx_friends_user_status ON Friends(User_Id, Status);
   CREATE INDEX IF NOT EXISTS idx_friends_friend_status ON Friends(Friend_Id, Status);
   ```

6. **Progress_History table:**
   ```sql
   CREATE INDEX IF NOT EXISTS idx_progress_history_user_date ON Progress_History(User_Id, Created_At);
   ```

### Query Optimization Tips

1. **Use LIMIT and OFFSET for pagination** - Already implemented
2. **Use prepared statements** - Already implemented via parameterized queries
3. **Avoid N+1 queries** - Use JOINs where appropriate
4. **Cache frequently accessed data** - Redis cache already implemented
5. **Use WAL mode** - Already enabled in main.go

### Connection Pool Settings

Current settings (optimized for 50-100 concurrent users):
- MaxOpenConns: 25
- MaxIdleConns: 5
- ConnMaxLifetime: 5 minutes
- ConnMaxIdleTime: 1 minute

These settings ensure efficient connection reuse while preventing connection exhaustion.

