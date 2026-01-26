# Deployment Visibility Improvements

## Overview

Enhance F8A's deployment visibility through real-time progress tracking, historical deployment data, and improved status indicators. The primary goals are to help users understand deployment progress in real-time and provide actionable information when things go wrong.

## Problem Statement

Current issues with deployment visibility:
- **Hard to see progress**: Scrolling terminal output makes it difficult to tell how far along a deployment is or if it's stuck
- **Missing actionable info**: When deployments fail, unclear what went wrong or what to do next
- **No deployment history**: Can't track past deployments or compare current vs deployed revisions
- **Workflow pattern**: Users typically work on one app with iterative deployments (multiple deploys per session)

## Architecture

### 1. Data Model - Deployment History Storage

Add new `deployments` table to SQLite database:

```sql
CREATE TABLE deployments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id INTEGER NOT NULL,
    commit_hash TEXT NOT NULL,
    status TEXT NOT NULL, -- 'running', 'success', 'failed'
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    total_duration_ms INTEGER,
    error_message TEXT,
    error_step TEXT,
    FOREIGN KEY (app_id) REFERENCES apps(id)
);

CREATE TABLE deployment_steps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    deployment_id INTEGER NOT NULL,
    step_name TEXT NOT NULL, -- 'build', 'archive', 'upload', etc.
    status TEXT NOT NULL, -- 'pending', 'running', 'success', 'failed'
    started_at DATETIME,
    completed_at DATETIME,
    duration_ms INTEGER,
    output TEXT, -- captured command output
    FOREIGN KEY (deployment_id) REFERENCES deployments(id)
);
```

### 2. Progress Tracking System

Create new `progress` package with:

**DeploymentProgress struct:**
```go
type DeploymentProgress struct {
    DeploymentID   int
    CurrentStep    string
    Steps          []Step
    StartTime      time.Time
    UpdateChannel  chan StepUpdate
}

type Step struct {
    Name           string
    Description    string
    Status         StepStatus // pending, running, success, failed
    StartTime      time.Time
    Duration       time.Duration
    AverageDuration time.Duration // from historical data
    Output         []string
}

type StepStatus int
const (
    StepPending StepStatus = iota
    StepRunning
    StepSuccess
    StepFailed
)
```

**Deployment steps sequence:**
1. Build - Run `yarn build` locally
2. Archive - Create tar file of dist folder
3. Upload - SCP transfer to remote server
4. Backup - Move current version to previous
5. Extract - Extract new build on server
6. Activate - Create symlink to new version
7. Cleanup - Remove tar file, copy .htaccess if needed

### 3. UI Components

**Three new components:**

1. **StatusCard** - Shows at top of app detail view
   - Current deployment status badge
   - Deployed revision vs local revision comparison
   - Quick action button (Deploy/Re-deploy)

2. **ProgressPanel** - Replaces/enhances terminal output during deployment
   - Step checklist with icons (spinner/checkmark/X)
   - Real-time timing per step
   - Historical average timing
   - Overall progress bar
   - Collapsible terminal output per step

3. **HistoryTimeline** - Below action buttons
   - Recent deployments list (last 10)
   - Expandable details per deployment
   - Retry button for failed deployments

## Detailed Design

### Status Card Component

Located at top of app detail view (components/app_content.go):

**Visual layout:**
```
┌─────────────────────────────────────────────────────────┐
│ Status: 🟢 Up to date                                   │
│ Remote: abc1234 (deployed 2h ago)                       │
│ Local:  abc1234 (current)                               │
│                                    [Re-deploy] button    │
└─────────────────────────────────────────────────────────┘
```

**Status indicators:**
- 🟢 Green "Up to date" - remote and local revisions match
- 🟡 Yellow "Local changes" - local is ahead of remote
- 🔴 Red "Deploy failed" - last deployment failed
- 🔵 Blue "Deploying..." - deployment in progress

**Logic:**
- Compare local git HEAD with remote revision on component load
- Update badge color and text based on comparison
- Show commit count difference if local is ahead ("+3 commits")
- Button text changes: "Deploy Latest" (when ahead), "Re-deploy" (when current)

### Progress Panel Component

Replaces current terminal output area during deployments:

**Visual layout during deployment:**
```
┌─────────────────────────────────────────────────────────┐
│ Deploying revision def5678                              │
│ ━━━━━━━━━━━━━━━━━━━━░░░░░░░░░░░░░░░░ 60% (1m 23s)      │
│                                                          │
│ ✓ Build             28s (avg: 25s)                      │
│ ✓ Archive           3s  (avg: 2s)                       │
│ ⟳ Upload            45s (avg: 12s)  ← currently running │
│   Backup            --  (avg: 1s)                       │
│   Extract           --  (avg: 5s)                       │
│   Activate          --  (avg: 1s)                       │
│   Cleanup           --  (avg: 2s)                       │
│                                                          │
│ [Show detailed output ▼]                                │
└─────────────────────────────────────────────────────────┘
```

**Icons:**
- ⟳ Spinner - step currently running
- ✓ Checkmark - step completed successfully
- ✗ X mark - step failed
- -- Dash - step pending

**Time display:**
- Elapsed time for completed/running steps
- Average time from last 10 successful deployments
- Overall progress percentage based on weighted steps
- Total elapsed time

**Expandable output:**
- "Show detailed output" toggle
- Expands to show full terminal output (current black background area)
- Per-step output can be collapsed/expanded individually

### History Timeline Component

Shows below action buttons, collapsible section:

**Visual layout:**
```
┌─────────────────────────────────────────────────────────┐
│ Recent Deployments ▼                                    │
│                                                          │
│ ✓ abc1234  2 hours ago    (1m 23s)  [View details]     │
│ ✓ def5678  5 hours ago    (1m 18s)  [View details]     │
│ ✗ ghi9012  1 day ago      failed at Upload             │
│                                      [Retry] [Details]  │
│ ✓ jkl3456  1 day ago      (1m 31s)  [View details]     │
│                                                          │
│ [View all history...]                                   │
└─────────────────────────────────────────────────────────┘
```

**Each deployment entry shows:**
- Status icon (✓ success, ✗ failed)
- Commit hash (clickable to show full hash)
- Relative timestamp ("2 hours ago")
- Duration or failure step
- Action buttons

**Expanded detail view (dialog):**
```
┌─────────────────────────────────────────────────────────┐
│ Deployment: abc1234                                     │
│ Started: 2026-01-26 12:34:56                           │
│ Duration: 1m 23s                                        │
│ Status: Success                                         │
│                                                          │
│ Steps:                                                  │
│ ✓ Build      28s                                        │
│ ✓ Archive    3s                                         │
│ ✓ Upload     12s                                        │
│ ✓ Backup     1s                                         │
│ ✓ Extract    5s                                         │
│ ✓ Activate   1s                                         │
│ ✓ Cleanup    2s                                         │
│                                                          │
│ [View output logs] [Close]                              │
└─────────────────────────────────────────────────────────┘
```

### Desktop Notifications

Use Fyne's notification support (fyne.Notification):

**On success:**
- Title: "F8A Deployment Complete"
- Body: "MyApp (abc1234) deployed successfully in 1m 23s"
- Icon: Success/checkmark icon

**On failure:**
- Title: "F8A Deployment Failed"
- Body: "MyApp deployment failed at Upload step"
- Icon: Error/X icon
- Action: Click to open app and show error details

**Notification timing:**
- Only send if app window is not focused
- Send immediately on deployment completion/failure
- Include sound alert for failures

## Implementation Plan

### Phase 1: Database & Models
1. Create database migration for new tables (deployments, deployment_steps)
2. Add model functions in `models/deployment.go`:
   - CreateDeployment, UpdateDeployment, GetDeployments
   - CreateDeploymentStep, UpdateDeploymentStep
   - GetAverageStepDuration (for time estimates)

### Phase 2: Progress Tracking
1. Create `progress/tracker.go` package:
   - DeploymentProgress struct and methods
   - Step execution wrapper that tracks timing
   - Channel-based UI updates
2. Refactor `commands/deploy.go` to use progress tracker:
   - Wrap each deployment step
   - Emit progress updates
   - Store step results to database

### Phase 3: UI Components
1. Create `components/status_card.go`:
   - Revision comparison logic
   - Status badge display
   - Quick deploy button
2. Create `components/progress_panel.go`:
   - Step checklist display
   - Progress bar
   - Time estimates
   - Expandable output
3. Create `components/history_timeline.go`:
   - Recent deployments list
   - Expandable detail view
   - Retry button functionality

### Phase 4: Notifications
1. Add notification support in `utils/notifications.go`:
   - Success notification
   - Failure notification
   - Window focus detection

### Phase 5: Integration
1. Update `components/app_content.go`:
   - Add StatusCard to top of app detail view
   - Replace output container with ProgressPanel
   - Add HistoryTimeline below action buttons
2. Update deployment flow to populate history on completion
3. Add revision checking on app selection

## Technical Considerations

### Performance
- Database queries for history: index on (app_id, started_at)
- Limit history display to last 10 by default
- Use goroutines for progress updates to avoid blocking UI

### Error Handling
- Store full error messages and step context in database
- Preserve terminal output for debugging
- Allow retry without losing error context

### Backward Compatibility
- Database migration is additive (no breaking changes)
- Existing apps continue to work
- History only available for new deployments

### Testing
- Test with slow network connections (upload step)
- Test with build failures
- Test with SSH connection issues
- Verify notification behavior when app is backgrounded

## Success Metrics

Implementation is successful if:
1. Users can immediately see which step a deployment is on
2. Time estimates help users know if deployment is progressing normally
3. Failed deployments clearly show what went wrong and where
4. Deployment history provides useful context for troubleshooting
5. Status card eliminates need to manually check git revisions
6. Desktop notifications catch attention when deployments complete

## Future Enhancements (Out of Scope)

- Deployment comparison (diff between two deployments)
- Export deployment history to CSV/JSON
- Slack/webhook notifications
- Parallel deployments to multiple servers
- Automatic rollback on failure detection
- Pre-deployment health checks
- Post-deployment smoke tests
