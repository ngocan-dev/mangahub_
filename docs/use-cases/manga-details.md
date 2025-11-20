# Manga Details Access (Manga Reader)

**Primary Actor:** Manga Reader  \
**Goal:** Access detailed information about a specific manga  \
**Preconditions:** Manga exists in the database  \
**Postconditions:** Complete manga information is displayed

## Main Success Scenario
1. User selects a manga from search results or navigates directly via URL.
2. System retrieves the manga’s details from the database.
3. System displays the title, author, genres, description, and total chapter count.
4. If the user is logged in, the system also shows the user’s current reading progress.
5. The user can add the manga to their library or update their reading progress from the detail view.

## Notes
- The flow assumes the database is reachable and returns a matching record; otherwise, the user should see a clear “manga not found” message.
- When the user is not authenticated, the details remain visible but library and progress actions should prompt for sign-in.
- Progress and library actions should validate chapter numbers and enforce library membership rules before updates are applied.
