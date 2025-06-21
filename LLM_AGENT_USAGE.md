# GTD Usage Guide for LLM Agents

This guide explains how to effectively use the `gtd` task management tool as an AI agent helping with software development and project management.

## Core Workflow Philosophy

The `gtd` tool follows the GTD (Getting Things Done) methodology with true INBOX-first capture:

1. **CAPTURE**: ALL new tasks go to INBOX (regardless of source or clarity)
2. **ENGAGE**: Work on NEW and IN_PROGRESS tasks first (your committed work)
3. **CLARIFY**: Review INBOX items only when current work is manageable  
4. **ORGANIZE**: Accept (move to NEW) or reject (mark INVALID) INBOX items
5. **REFLECT**: Regular reviews and planning

**Key Principle:** Everything starts in INBOX. Complete your current commitments before reviewing INBOX and accepting new work.

## Essential Commands for AI Agents

### Task Creation (CAPTURE)

```bash
# ALL tasks go to INBOX first for review - no matter how clear they seem
echo "Memory leak in user session handling
Found during load testing - heap grows indefinitely
Affects production servers running >24 hours" | gtd add bug --priority high --source "monitoring:alert-123"

# Features go to INBOX for prioritization decisions
echo "Implement OAuth2 authentication
Support Google and GitHub providers
Include JWT token refresh mechanism" | gtd add feature --priority medium --tags "auth,security"

# Regressions go to INBOX for urgency assessment
echo "Search functionality broken after API refactor
Query filters no longer work properly" | gtd add regression --source "git:abc123def" --priority high

# Everything gets captured in INBOX - no exceptions
echo "Investigate user complaint about slow response" | gtd add bug --source "support:ticket-456"
```

**Key principles:**
- ALL add commands create tasks in INBOX state
- Always provide detailed descriptions with context
- Include reproduction steps for bugs
- Specify business impact and urgency
- Use meaningful tags for categorization
- Reference sources (git commits, issues, monitoring alerts)
- Everything gets reviewed later - capture first, decide during review

### Review and Triage (CLARIFY/ORGANIZE)

```bash
# First, check your current workload - focus on committed work
gtd list  # See active tasks (NEW, IN_PROGRESS)

# Only review INBOX when current work is manageable
gtd review  # Tool will warn if you have too many active tasks

# For each task in INBOX, decide its fate:
gtd show 1a2b3c4  # Get full details

# Accept task (move to NEW state, becomes committed work)
gtd accept 1a2b3c4

# Reject task (mark as INVALID)
gtd reject 1a2b3c4  # Use for duplicates, invalid reports, out-of-scope items
```

**Review criteria:**
- Is the task clearly defined and actionable?
- Is it aligned with current project goals?
- Is it a duplicate of existing work?
- Does it provide enough context to work on?
- Is the priority and effort estimate realistic?
- Do you have capacity to commit to this work?

**Critical principle:** Accepting a task means committing to do it. Only accept what you can realistically complete.

### Task Planning and Organization

```bash
# Break down large tasks into subtasks
echo "Set up CI/CD pipeline" | gtd add-subtask 1a2b3c4 --kind feature --priority high
echo "Write integration tests" | gtd add-subtask 1a2b3c4 --kind feature --priority medium
echo "Deploy to staging" | gtd add-subtask 1a2b3c4 --kind feature --priority medium

# Establish dependencies between tasks
gtd block 5e6f7g8 --by 1a2b3c4  # Task 5e6f7g8 is blocked by 1a2b3c4

# Check blocking relationships
gtd show 5e6f7g8  # Will show what tasks are blocking this one
```

### Daily Workflow Management

```bash
# Start your day: honor your commitments first
gtd list --priority high  # See high-priority committed tasks
gtd list --state in_progress  # Continue work in progress
gtd list --blocked  # Check for unblocked tasks

# During work: update task states
gtd in-progress 1a2b3c4  # Start working on a task
gtd done 1a2b3c4  # Complete a task
gtd cancel 1a2b3c4  # Cancel if no longer needed

# Mid-day: check if you can take on more work
gtd list  # Review your current load

# End of day: review progress and process new items
gtd summary  # Get overview of all tasks
gtd list --state in_progress  # See what's currently in progress

# Only when current commitments are under control:
gtd review  # Process INBOX items and make new commitments
```

### Information Gathering and Reporting

```bash
# Find specific tasks
gtd search "memory leak"  # Search in titles and descriptions
gtd list --kind bug --priority high  # Filter by type and priority
gtd list --tag security  # Find all security-related tasks

# Generate reports
gtd summary  # Overall project status
gtd export --format markdown --state done --output weekly-report.md  # Export completed tasks
gtd export --format json --all  # Full data export for analysis
```

## Best Practices for AI Agents

### 1. Task Creation Standards

**Always include:**
- Clear, specific title (under 80 characters)
- Detailed description with context
- Steps to reproduce (for bugs)
- Acceptance criteria (for features)
- Business impact and urgency
- Relevant tags and source references

**Example of good task creation:**
```bash
echo "API rate limiting returns 500 instead of 429
When API rate limit is exceeded (>1000 req/min), server returns 500 Internal Server Error instead of proper 429 Too Many Requests.

Steps to reproduce:
1. Send >1000 requests/minute to /api/users
2. Observe response code

Expected: HTTP 429 with rate limit headers
Actual: HTTP 500 with generic error message

Impact: Poor client experience, difficult debugging" | gtd add bug --priority high --tags "api,rate-limiting" --source "logs:2024-01-15"
```

### 2. Regular Review Cycles

**Daily:**
```bash
# Morning routine - focus on commitments
gtd list --priority high  # Plan day's work with committed tasks
gtd list --state in_progress  # Continue work in progress

# Evening routine - process and commit
gtd summary  # Review progress on commitments
gtd list --state in_progress  # Check what's still active
gtd review  # Process INBOX only if you can handle more work
```

**Weekly:**
```bash
# Project health check
gtd export --format markdown --output weekly-status.md
gtd list --blocked  # Address blocking issues
gtd list --kind bug --priority high  # Critical bug review
gtd review  # Major INBOX processing session
```

### 3. Effective Filtering and Searching

```bash
# Find tasks by context
gtd search "authentication"  # Find all auth-related work
gtd list --tag "security" --priority high  # Security priorities
gtd list --kind bug --state new  # New bugs to triage

# Monitor progress
gtd list --state in_progress  # What's currently being worked on
gtd list --kind feature --priority high  # High-priority features
```

### 4. Dependency Management

```bash
# When creating dependent tasks
echo "Deploy new API version" | gtd add-feature --priority medium
echo "Update client library" | gtd add-feature --priority medium  
gtd block 2b3c4d5 --by 1a2b3c4  # Client update blocked by API deployment

# Regular dependency review
gtd list --blocked  # See what's waiting on other work
gtd show 1a2b3c4  # Check what this task is blocking
```

## Integration with Development Workflow

### Code Changes to Task Updates

When making code changes, update related tasks:

```bash
# Starting work on a task
gtd in-progress abc123  # Mark as in progress
# ... make code changes ...
gtd done abc123  # Mark as complete when done

# If you discover new issues while working
echo "Discovered memory leak in user session cleanup
Found while implementing OAuth2 - sessions not properly cleared on logout" | gtd add-bug --priority medium --source "task:abc123"
```

### Issue Tracking Integration

```bash
# Reference external systems
gtd add bug --source "github:issue/123"
gtd add bug --source "jira:BUG-456" 
gtd add bug --source "slack:C123/p456789"
gtd add bug --source "sentry:error-789"
```

## Common Patterns and Workflows

### Bug Report Processing

```bash
# 1. Capture ALL bug reports in INBOX
echo "$BUG_REPORT_CONTENT" | gtd add bug --priority medium --source "user-report:123"

# 2. During review session, triage
gtd review
gtd show 1a2b3c4  # Get full details

# 3. Accept if it's valid and actionable
gtd accept 1a2b3c4  # Now it becomes committed work

# 4. Start investigation
gtd in-progress 1a2b3c4
# ... after investigation ...
# If critical: create subtasks for immediate fix (also go to INBOX first)
echo "Hotfix for critical login bug" | gtd add-subtask 1a2b3c4 --kind bug --priority high
```

### Feature Development Lifecycle

```bash
# 1. Capture ALL feature requests in INBOX (even well-defined ones)
echo "User dashboard with real-time metrics
Allow users to see their activity in real-time
Include charts for data visualization" | gtd add feature --priority medium --tags "ui,dashboard"

# 2. During review, evaluate and accept
gtd review
gtd accept 1a2b3c4  # Accept if aligned with goals and capacity

# 3. Break down into subtasks (these also go to INBOX)
echo "Design dashboard mockups" | gtd add-subtask 1a2b3c4 --kind feature --priority high
echo "Implement real-time data API" | gtd add-subtask 1a2b3c4 --kind feature --priority high  
echo "Create dashboard frontend" | gtd add-subtask 1a2b3c4 --kind feature --priority medium
echo "Write integration tests" | gtd add-subtask 1a2b3c4 --kind feature --priority medium

# 4. Accept subtasks during next review
gtd review
gtd accept 2b3c4d5  # Accept API task
gtd accept 3c4d5e6  # Accept frontend task

# 5. Establish dependencies and execute
gtd block 3c4d5e6 --by 2b3c4d5  # Frontend blocked by API
gtd in-progress 2b3c4d5  # Start with API
```

### Sprint Planning

```bash
# Review current commitments first
gtd summary
gtd list --state new --priority high  # High-priority committed work
gtd list --state in_progress  # Work currently being done

# Process INBOX to identify new candidates
gtd review
# Accept items that fit sprint capacity and goals
gtd accept 1a2b3c4  # Accept high-priority bug
gtd accept 5e6f7g8  # Accept important feature
gtd reject 9g0h1i2  # Reject out-of-scope item

# Plan sprint with newly committed work
gtd list --state new --kind bug  # Bugs to address in sprint
gtd list --state new --kind feature --priority high  # Features for sprint

# Track sprint progress
gtd list --state in_progress  # What's currently being worked on
gtd export --format json --state done  # Completed work for review
```

## Error Handling and Recovery

### When Tasks Are Unclear

```bash
# If a task in INBOX lacks detail
gtd show 1a2b3c4  # Review what we have
gtd reject 1a2b3c4  # Reject if insufficient info

# Create a new, clearer task when you have the details
echo "Need more details for original bug report
Original report lacks reproduction steps and context" | gtd add bug --priority low --source "task:1a2b3c4"
```

### When Priorities Change

```bash
# Find tasks that need priority updates
gtd list --priority low  # Review low-priority items
gtd search "critical"  # Find tasks that might need urgent attention

# Use export and external tools for bulk updates when needed
gtd export --format json | jq '.' | # process with external tools
```

## Performance Tips

- Use `--oneline` for quick scans: `gtd list --oneline`
- Filter early and often: `gtd list --kind bug --priority high`
- Export for complex analysis: `gtd export --format json`
- Regular cleanup: review and remove/cancel obsolete tasks

## Integration with AI Development Workflows

As an AI agent, you can:

1. **Automatically capture ALL issues** in INBOX for later review
2. **Track task dependencies** when planning feature development  
3. **Generate reports** for human developers
4. **Maintain task hygiene** by regular INBOX reviews and commitment management
5. **Correlate tasks with code changes** using source references
6. **Enforce the INBOX-first workflow** by ensuring nothing bypasses the review process

Remember: The goal is maintaining a trusted system where everything gets captured in INBOX, reviewed deliberately, and only committed work appears in the active task list. This helps human developers focus on execution while ensuring nothing gets lost.