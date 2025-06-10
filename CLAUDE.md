# Budget Tracker - AI Assistant Guide

## Project Overview

This is a terminal-based personal finance tool built with Go, designed for fast keyboard-driven interaction. The application provides comprehensive budget tracking with multi-currency support and real-time financial insights.

## Architecture

### Technology Stack
- **Language**: Go 1.24+
- **UI Framework**: Bubble Tea (charm.sh/bubbletea)
- **Database**: SQLite with GORM ORM
- **Testing**: Real SQLite databases, no mocks
- **Pattern**: Repository pattern with service layer

### Project Structure
```
budget-tracker/
├── cmd/budget/         # Application entry point
├── internal/
│   ├── models/        # Data structures
│   ├── repository/    # Database access layer
│   ├── service/       # Business logic
│   ├── ui/           # Terminal UI components
│   └── db/           # Database initialization
├── test/             # Test utilities and fixtures
└── data/             # SQLite database storage
```

## Key Design Decisions

### 1. Real Database Testing
- Each test uses its own SQLite database file
- No mocks - tests verify actual functionality
- Automatic cleanup after tests
- Skip API tests with `-short` flag

### 2. Multi-Currency Architecture
- All amounts stored in original currency
- USD as base currency for aggregation
- Exchange rates cached for 1 hour
- Fixed rates for pegged currencies (AED)

### 3. Keyboard-First UI
- Single-key shortcuts for all major actions
- Modal dialogs for data entry
- Vim-like navigation (j/k for up/down)
- No mouse interaction required

## Common Tasks

### Adding a New Model
1. Create model in `internal/models/`
2. Add to database migrations in `internal/db/init.go`
3. Create repository in `internal/repository/`
4. Add service methods in `internal/service/`
5. Write tests in repository test file

### Adding a UI View
1. Create new view in `internal/ui/views/`
2. Implement `tea.Model` interface
3. Add navigation in main model
4. Register keyboard shortcuts
5. Test keyboard navigation

### Currency Integration
1. Check if currency has fixed rate
2. If not, fetch from ExchangeRate-API
3. Cache rate for 1 hour
4. Convert to USD for storage
5. Display in original currency

## Keyboard Shortcuts Reference

### Global
- `q` - Quit application
- `Esc` - Cancel/back
- `Tab` - Next field
- `Shift+Tab` - Previous field
- `/` - Quick search

### Main Dashboard
- `n` - New transaction
- `t` - View all transactions
- `b` - Manage budgets
- `r` - View reports
- `c` - Manage categories

### List Views
- `j`/`↓` - Move down
- `k`/`↑` - Move up
- `Enter` - Select/edit
- `d` - Delete (with confirmation)
- `f` - Filter options

### Transaction Entry
- `Tab` - Navigate fields
- `Enter` - Save transaction
- `Esc` - Cancel without saving

## Testing Approach

### Unit Tests
```go
func TestTransactionCreate(t *testing.T) {
    db := test.SetupTestDB(t)
    repo := repository.NewTransactionRepository(db)
    
    tx := &models.Transaction{
        Type: "expense",
        Amount: 50.00,
        Currency: "USD",
    }
    
    err := repo.Create(tx)
    require.NoError(t, err)
}
```

### Integration Tests
- Test complete workflows
- Verify UI updates correctly
- Check budget calculations
- Test currency conversions

### Running Tests
```bash
make test          # Run all tests
make test-short    # Skip API tests
make test-cover    # Generate coverage report
```

## Known Limitations

1. **Exchange Rate API**: Limited to 1000 requests/month on free tier
2. **Currency Support**: Only major currencies + AED
3. **Terminal Only**: No web interface
4. **Single User**: No multi-user support
5. **Local Storage**: SQLite database stored locally

## Performance Considerations

- Database indexes on frequently queried fields
- Lazy loading for transaction lists
- Cached exchange rates
- Pagination for large datasets
- Efficient terminal rendering

## Error Handling

- User-friendly error messages
- No stack traces in UI
- Log errors to file for debugging
- Graceful fallbacks for API failures
- Database transaction rollbacks

## Development Workflow

1. Write failing test first
2. Implement minimal code to pass
3. Refactor for clarity
4. Ensure UI integration works
5. Update documentation if needed

## Debugging Tips

- Check `data/budget.log` for errors
- Use `BUDGET_DEBUG=1` for verbose output
- SQLite browser for database inspection
- Test individual components in isolation
- Use `tea.Printf` for UI debugging