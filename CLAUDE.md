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

## Upcoming Features Architecture

### 1. Editable Categories (UI-Based Approach)

**Rationale**: UI-based editing provides better control over edge cases and user experience.

**Implementation Plan**:
- **Category Management View**: New UI view accessible via 'c' from dashboard
- **Features**:
  - List all categories with usage counts
  - Edit category name, icon, and color
  - Create custom categories
  - Delete unused categories only
  - Merge categories with transaction migration
- **Edge Case Handling**:
  - Cannot delete categories with existing transactions
  - Merge operation: transfers all transactions to target category
  - Maintain audit trail of category changes
  - Prevent duplicate category names within same type

**Database Changes**:
```sql
-- Add category history table
CREATE TABLE category_history (
    id INTEGER PRIMARY KEY,
    category_id INTEGER NOT NULL,
    old_name VARCHAR(100),
    new_name VARCHAR(100),
    action VARCHAR(20), -- 'rename', 'merge', 'delete'
    target_category_id INTEGER,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(id)
);
```

### 2. Configurable Currencies

**Settings File**: `data/settings.json`

```json
{
  "currencies": {
    "enabled": ["USD", "EUR", "AED"],
    "default": "USD",
    "fixed_rates": {
      "AED": 3.6725
    }
  },
  "ui": {
    "date_format": "2006-01-02",
    "decimal_places": 2
  }
}
```

**Implementation**:
- **Settings Service**: Load/save configuration
- **Currency Management View**: Toggle currencies on/off
- **Validation**: Prevent disabling currencies with existing transactions
- **Migration**: Convert disabled currency transactions to default currency

### 3. Recurring Transactions

**Model Structure**:
```go
type RecurringTransaction struct {
    ID              uint
    Type            TransactionType
    Amount          float64
    Currency        string
    CategoryID      uint
    Description     string
    Frequency       RecurrenceFrequency // daily, weekly, monthly, yearly
    FrequencyValue  int                 // e.g., every 2 weeks
    StartDate       time.Time
    EndDate         *time.Time
    LastProcessed   *time.Time
    IsActive        bool
    NextDueDate     time.Time
    // Relationships
    Category        Category
    Transactions    []Transaction // Generated transactions
}

type RecurrenceFrequency string
const (
    FrequencyDaily   RecurrenceFrequency = "daily"
    FrequencyWeekly  RecurrenceFrequency = "weekly"
    FrequencyMonthly RecurrenceFrequency = "monthly"
    FrequencyYearly  RecurrenceFrequency = "yearly"
)
```

**Features**:
- **Automatic Generation**: Daily job to create due transactions
- **Manual Override**: Edit/skip individual occurrences
- **UI Management**: List, create, edit, pause recurring transactions
- **Smart Detection**: Suggest recurring patterns from transaction history

### 4. Yearly Projections

**Projection Algorithm**:
1. Calculate average monthly expenses from last 3-6 months
2. Include all active recurring transactions
3. Factor in seasonal variations (if data available)
4. Project forward 12 months

**Report Enhancements**:
- **Projection View**: New section in reports showing:
  - Projected monthly expenses
  - Projected yearly total
  - Budget vs projection comparison
  - Confidence level based on data consistency
- **Visualization**: ASCII charts showing trends

**API Structure**:
```go
type YearlyProjection struct {
    Year                int
    MonthlyProjections  []MonthProjection
    TotalProjected      float64
    RecurringTotal      float64
    VariableTotal       float64
    ConfidenceLevel     float64 // 0-1
}

type MonthProjection struct {
    Month           time.Month
    ProjectedAmount float64
    RecurringAmount float64
    VariableAmount  float64
    BudgetAmount    float64
}
```

## Implementation Priority

1. **Phase 1: Settings & Currency Configuration** ✅ COMPLETED
   - Settings service and JSON configuration
   - Currency enable/disable UI
   - Migration logic for currency changes

2. **Phase 2: Category Management** (NEXT)
   - Category management UI view
   - Edit/merge functionality
   - Transaction migration on category changes
   - History tracking

3. **Phase 3: Recurring Transactions**
   - Data model and migrations
   - Service layer for processing
   - UI for management
   - Automatic transaction generation

4. **Phase 4: Projections & Enhanced Reports**
   - Projection calculation service
   - Enhanced reports view
   - Visualization components

## Migration Strategy

1. **Database Migrations**: Version-controlled migrations
2. **Data Safety**: Backup before major changes
3. **Rollback Plan**: Each feature can be disabled via settings
4. **Testing**: Comprehensive tests for edge cases

## Recent Changes

### Currency Configuration (Completed)
- Added `SettingsService` for managing application configuration
- Created currency settings UI accessible via 'u' from dashboard
- Implemented enable/disable functionality with validation
- Settings stored in `data/settings.json` with automatic creation
- Updated `CurrencyService` to use settings-based enabled currencies
- Fixed rates now stored in settings (e.g., AED = 3.6725)
- Added protection against disabling currencies with existing transactions
- Thread-safe concurrent access to settings

### Key Components Added
1. **Models**:
   - `models.Settings`: Application configuration structure
   - `models.CategoryHistory`: For tracking category changes

2. **Services**:
   - `SettingsService`: Manages JSON configuration file
   - Updated `CurrencyService`: Now uses settings for enabled currencies

3. **UI Views**:
   - `CurrencySettings`: Interactive currency management interface

4. **Integration Points**:
   - Main app initialization includes settings service
   - All services updated to accept settings service
   - Dashboard updated with 'u' shortcut for currency settings