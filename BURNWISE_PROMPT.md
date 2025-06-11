# BurnWise Implementation Prompt

You are tasked with creating a standalone burn rate monitoring application in Go. This will be a terminal-based expense tracking tool focused on monitoring monthly burn rate, recurring expenses, and financial projections.

## Project Requirements

### Core Functionality
Build a FULLY FUNCTIONAL budget tracker where:
- EVERY feature mentioned in the UI actually works - no TODO stubs
- Users can track income and expenses with categories
- Multi-currency support with USD as base currency
- Monthly and yearly budget planning
- Real-time budget vs actual tracking
- Transaction history with search/filter
- Financial reports and summaries

### Technical Stack
- **Language**: Go 1.24+
- **UI**: Bubble Tea (terminal UI framework)
- **Database**: SQLite with GORM
- **Testing**: Pragmatic TDD with real SQLite databases
- **Architecture**: Repository pattern with service layer

## Project Structure
```
budget-tracker/
├── cmd/
│   └── budget/
│       └── main.go          # Entry point
├── internal/
│   ├── models/              # Data models
│   ├── repository/          # Database access
│   ├── service/             # Business logic
│   ├── ui/                  # Terminal UI
│   └── db/                  # Database initialization
├── test/
│   ├── fixtures/            # Test data builders
│   └── integration/         # End-to-end tests
├── data/                    # SQLite database location
├── CLAUDE.md               # AI assistance guide
├── TASKS.md                # Current progress tracking
├── TDD.md                  # Test strategy document
├── README.md               # User documentation
├── Makefile                # Build commands
└── go.mod
```

## Data Models

### Transaction
```go
type Transaction struct {
    ID           uint
    Type         TransactionType  // "income", "expense", "transfer"
    Amount       float64         // In original currency
    Currency     string          // ISO code (USD, EUR, AED)
    AmountUSD    float64         // Converted to USD for aggregation
    CategoryID   uint
    Description  string
    Date         time.Time
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    gorm.DeletedAt

    // Relationships
    Category     Category
}
```

### Category
```go
type Category struct {
    ID           uint
    Name         string
    Type         TransactionType
    Icon         string          // Emoji for display
    Color        string          // Hex color
    ParentID     *uint          // For subcategories
    IsDefault    bool           // System categories
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    gorm.DeletedAt

    // Relationships
    Parent       *Category
    Transactions []Transaction
}
```

### Budget
```go
type Budget struct {
    ID           uint
    Name         string
    CategoryID   uint
    Amount       float64         // Monthly limit in USD
    Period       string          // "monthly", "yearly"
    StartDate    time.Time
    EndDate      *time.Time      // NULL for ongoing
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    gorm.DeletedAt

    // Relationships
    Category     Category
}
```

## UI Requirements

### Main Dashboard View
```
🔥 BurnWise                                          October 2025

Income:    $5,000.00    ████████████████████ 100%
Expenses:  $3,500.00    ██████████████       70%
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Balance:   $1,500.00

Recent Transactions
Date        Category        Description          Amount
─────────────────────────────────────────────────────────
10/15       🍔 Food        Lunch at cafe        -$25.00
10/15       💼 Salary      Monthly salary     +$5,000.00
10/14       🏠 Rent        October rent       -$1,500.00

[n]ew  [t]ransactions  [b]udgets  [r]eports  [q]uit
```

### Transaction Entry Modal
```
┌─────────────────────────────────────┐
│       Add Transaction               │
├─────────────────────────────────────┤
│ Type:     [Income ▼]               │
│ Amount:   [_______]                 │
│ Currency: [USD ▼]                  │
│ Category: [Select... ▼]            │
│ Description: [__________________]  │
│ Date:     [2025-10-15_]           │
│                                     │
│        [Save]  [Cancel]             │
└─────────────────────────────────────┘
```

### Keyboard Navigation
- `n` - New transaction
- `t` - View all transactions
- `b` - Manage budgets
- `r` - View reports
- `e` - Edit selected
- `d` - Delete selected (with confirmation)
- `f` - Filter/search
- `/` - Quick search
- `Tab` - Navigate fields
- `Enter` - Confirm
- `Esc` - Cancel/back
- `q` - Quit

## Default Categories

### Income Categories
```go
var defaultIncomeCategories = []Category{
    {Name: "Salary", Icon: "💼", Type: "income"},
    {Name: "Freelance", Icon: "💻", Type: "income"},
    {Name: "Investments", Icon: "📈", Type: "income"},
    {Name: "Other Income", Icon: "💰", Type: "income"},
}
```

### Expense Categories
```go
var defaultExpenseCategories = []Category{
    {Name: "Housing", Icon: "🏠", Type: "expense"},
    {Name: "Food & Dining", Icon: "🍔", Type: "expense"},
    {Name: "Transportation", Icon: "🚗", Type: "expense"},
    {Name: "Shopping", Icon: "🛒", Type: "expense"},
    {Name: "Entertainment", Icon: "🎮", Type: "expense"},
    {Name: "Healthcare", Icon: "💊", Type: "expense"},
    {Name: "Education", Icon: "📚", Type: "expense"},
    {Name: "Utilities", Icon: "💡", Type: "expense"},
    {Name: "Other Expenses", Icon: "💸", Type: "expense"},
}
```

## Currency Handling

### Supported Currencies
- USD (base currency)
- EUR, GBP, JPY, CHF, CAD, AUD, NZD (via API)
- AED (fixed rate: 1 USD = 3.6725 AED)
- All amounts stored in original currency
- Exchange rates cached for 1 hour
- Use free ExchangeRate-API for conversion

### Fixed Rate Implementation
```go
// For AED and other pegged currencies
var fixedRates = map[string]float64{
    "AED": 3.6725,  // 1 USD = 3.6725 AED
}
```

## Testing Requirements

### Test Philosophy
1. Use REAL SQLite databases for all tests
2. Each test gets its own database file
3. Automatic cleanup after tests
4. Test actual functionality, not mocks
5. API tests can be skipped with -short flag

### Test Structure
```go
func TestCreateTransaction(t *testing.T) {
    // Setup
    db := test.SetupTestDB(t)
    repo := repository.NewTransactionRepository(db)
    
    // Create test transaction
    tx := &models.Transaction{
        Type:        "expense",
        Amount:      50.00,
        Currency:    "USD",
        CategoryID:  1,
        Description: "Test transaction",
        Date:        time.Now(),
    }
    
    // Test
    err := repo.Create(tx)
    require.NoError(t, err)
    assert.Greater(t, tx.ID, uint(0))
    
    // Verify
    found, err := repo.GetByID(tx.ID)
    require.NoError(t, err)
    assert.Equal(t, tx.Description, found.Description)
}
```

### Test Helpers
```go
// test/helpers/db.go
func SetupTestDB(t *testing.T) *gorm.DB {
    dbPath := fmt.Sprintf("./test/data/test_%s.db", t.Name())
    db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
    require.NoError(t, err)
    
    // Run migrations
    err = db.AutoMigrate(&models.Transaction{}, &models.Category{}, &models.Budget{})
    require.NoError(t, err)
    
    // Cleanup after test
    t.Cleanup(func() {
        sqlDB, _ := db.DB()
        sqlDB.Close()
        os.Remove(dbPath)
    })
    
    return db
}
```

## Documentation Requirements

### CLAUDE.md
Create a comprehensive guide for AI assistants including:
- Project overview and architecture
- Key design decisions
- Common tasks and workflows
- Testing approach
- Keyboard shortcuts reference
- Known limitations

### TASKS.md
Track implementation progress:
- [ ] Core models and database setup
- [ ] Transaction CRUD operations
- [ ] Category management
- [ ] Basic UI with table view
- [ ] Transaction entry modal
- [ ] Budget creation and tracking
- [ ] Monthly/yearly reports
- [ ] Multi-currency support
- [ ] Search and filtering
- [ ] Data export (CSV)

### TDD.md
Document the test-driven approach:
- Pragmatic testing philosophy
- Real database usage
- Test data fixtures
- Integration test examples
- How to run tests (make test, make test-short)
- Coverage goals (80%+ for business logic)

### README.md
User-friendly documentation:
- Installation instructions
- Features overview
- Usage examples
- Keyboard shortcuts
- Configuration options
- Screenshots (ASCII art)

## Implementation Priority

### Phase 1: Foundation (Week 1)
1. Set up project structure
2. Create all documentation files
3. Implement core models
4. Basic database operations
5. Simple transaction list view

### Phase 2: Core Features (Week 2)
1. Transaction CRUD with UI
2. Category management
3. Multi-currency support
4. Basic filtering

### Phase 3: Budgeting (Week 3)
1. Budget models and CRUD
2. Budget vs actual calculations
3. Visual budget progress
4. Overspending alerts

### Phase 4: Polish (Week 4)
1. Reports and analytics
2. Data export
3. Performance optimization
4. Comprehensive testing

## Critical Requirements

1. **NO TODO COMMENTS** - Implement everything fully
2. **Every UI element must work** - No placeholder buttons
3. **Test each feature** - Write a test that proves it works
4. **Commit frequently** - At least after each working feature
5. **User feedback** - Show clear error messages and confirmations

## Success Criteria

The budget tracker is complete when:
1. All CRUD operations work for transactions, categories, and budgets
2. Multi-currency conversions are accurate
3. Reports show correct calculations
4. All keyboard shortcuts function as documented
5. Tests pass and cover core functionality
6. Documentation is complete and accurate

## Example Workflow

```bash
# Start the app
./budget

# User presses 'n' to add transaction
# Modal appears, user enters:
# - Type: Expense
# - Amount: 50
# - Currency: AED
# - Category: Food & Dining
# - Description: Lunch at cafe
# - Date: today

# Transaction saved, appears in list
# Dashboard updates totals
# Budget progress reflects new expense
```

## Remember

You're building a tool for someone who:
- Values speed and keyboard efficiency
- Wants to track expenses in multiple currencies
- Needs clear visual feedback on budgets
- Expects everything to "just work"

Make it fast, make it reliable, make it complete.
