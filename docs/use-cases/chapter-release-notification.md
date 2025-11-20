# Chapter Release Notification Broadcast

**Primary Actor:** System Administrator  \
**Goal:** Notify users about new chapter releases  \
**Preconditions:** UDP server has registered clients  \
**Postconditions:** Notification is broadcasted to clients

## Main Success Scenario
1. Administrator triggers a notification for a specific manga and chapter.
2. System creates a notification message containing manga title, chapter number, and release notes if available.
3. UDP server broadcasts the notification to all registered clients.
4. Clients receive the notification and display it to end users.
5. System logs the successful broadcast.

## Alternative Flows
- **A1: Client unreachable**
  - Server continues broadcasting to other clients and logs the unreachable client for follow-up.
- **A2: Network error**
  - Server logs the error and retries the broadcast according to configured backoff limits.

## Notes
- Administrative trigger may come from a CLI command, dashboard action, or automated webhook connected to the chapter pipeline.
- Logged metadata should include manga identifier, chapter number, timestamp, and number of clients reached to support operational dashboards.
- Retries should avoid duplicate notifications by including a unique broadcast ID or idempotency key in log records.
