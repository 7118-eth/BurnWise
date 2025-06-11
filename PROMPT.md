# BurnWise - Phase 3 Implementation Prompt

I'm continuing work on a budget tracker application in Go. Please read the following documentation files to understand the current state:

1. Read CLAUDE.md (project guidelines and recent changes)
2. Read TASKS.md (to see completed phases and next tasks)
3. Read the compact context
4. Read FEATURES.md section on "In-Progress Features"

We've completed:
- Phase 1: Core budget tracking features
- Phase 1.5: Configurable currencies with settings
- Phase 2: Category management with edit/merge functionality

Now implement Phase 3: Recurring Transactions. Based on the architecture in CLAUDE.md:

1. Create the RecurringTransaction model in internal/models/
2. Update database migrations in internal/db/init.go
3. Create repository layer in internal/repository/
4. Implement service layer with:
   - Create/update/delete recurring transactions
   - Generate due transactions (daily job)
   - Handle skip/edit single occurrences
   - Smart pattern detection from history
5. Build UI views:
   - List view for recurring transactions
   - Form for creating/editing
   - Management of active/paused items
6. Integrate with main app (new view, keyboard shortcut)
7. Write comprehensive tests

Follow the established patterns, use real SQLite for tests, and maintain the keyboard-first UI approach. Start by reading the files, then create a todo list and begin implementation.