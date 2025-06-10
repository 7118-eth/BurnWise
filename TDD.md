# Test-Driven Development Strategy

## Philosophy

This project follows a pragmatic TDD approach focused on real-world functionality over mocking. We test against actual SQLite databases to ensure our code works in production-like conditions.

## Core Principles

1. **Real Databases Only**: No mocks, no stubs - test with actual SQLite
2. **Test First**: Write failing tests before implementation
3. **Isolated Tests**: Each test gets its own database
4. **Automatic Cleanup**: Tests clean up after themselves
5. **Fast Feedback**: Tests should run quickly (<5 seconds total)

## Test Structure

### Database Setup
```go
// test/helpers/db.go
package test

import (
    "fmt"
    "os"
    "testing"
    
    "github.com/stretchr/testify/require"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    
    "budget-tracker/internal/models"
)

func SetupTestDB(t *testing.T) *gorm.DB {
    t.Helper()
    
    // Create unique database for this test
    dbPath := fmt.Sprintf("./test/data/test_%s_%d.db", t.Name(), time.Now().UnixNano())
    
    // Ensure directory exists
    os.MkdirAll("./test/data", 0755)
    
    // Open database
    db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    require.NoError(t, err)
    
    // Run migrations
    err = db.AutoMigrate(
        &models.Transaction{},
        &models.Category{},
        &models.Budget{},
    )
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

### Test Fixtures
```go
// test/fixtures/transaction.go
package fixtures

import (
    "time"
    "budget-tracker/internal/models"
)

type TransactionBuilder struct {
    tx *models.Transaction
}

func NewTransaction() *TransactionBuilder {
    return &TransactionBuilder{
        tx: &models.Transaction{
            Type:        models.TransactionTypeExpense,
            Amount:      100.00,
            Currency:    "USD",
            AmountUSD:   100.00,
            Description: "Test transaction",
            Date:        time.Now(),
        },
    }
}

func (b *TransactionBuilder) WithAmount(amount float64) *TransactionBuilder {
    b.tx.Amount = amount
    b.tx.AmountUSD = amount // Assuming USD for simplicity
    return b
}

func (b *TransactionBuilder) WithCategory(categoryID uint) *TransactionBuilder {
    b.tx.CategoryID = categoryID
    return b
}

func (b *TransactionBuilder) Build() *models.Transaction {
    return b.tx
}
```

## Test Categories

### 1. Repository Tests
Test database operations in isolation:
```go
func TestTransactionRepository_Create(t *testing.T) {
    db := test.SetupTestDB(t)
    repo := repository.NewTransactionRepository(db)
    
    tx := fixtures.NewTransaction().Build()
    
    err := repo.Create(tx)
    require.NoError(t, err)
    assert.Greater(t, tx.ID, uint(0))
}

func TestTransactionRepository_GetByDateRange(t *testing.T) {
    db := test.SetupTestDB(t)
    repo := repository.NewTransactionRepository(db)
    
    // Create test data
    yesterday := fixtures.NewTransaction().
        WithDate(time.Now().AddDate(0, 0, -1)).
        Build()
    today := fixtures.NewTransaction().
        WithDate(time.Now()).
        Build()
    tomorrow := fixtures.NewTransaction().
        WithDate(time.Now().AddDate(0, 0, 1)).
        Build()
    
    repo.Create(yesterday)
    repo.Create(today)
    repo.Create(tomorrow)
    
    // Test date range
    results, err := repo.GetByDateRange(
        time.Now().AddDate(0, 0, -1),
        time.Now(),
    )
    
    require.NoError(t, err)
    assert.Len(t, results, 2)
}
```

### 2. Service Tests
Test business logic with real database:
```go
func TestBudgetService_CheckOverspending(t *testing.T) {
    db := test.SetupTestDB(t)
    budgetRepo := repository.NewBudgetRepository(db)
    txRepo := repository.NewTransactionRepository(db)
    service := service.NewBudgetService(budgetRepo, txRepo)
    
    // Create budget
    budget := &models.Budget{
        CategoryID: 1,
        Amount:     1000.00,
        Period:     "monthly",
        StartDate:  time.Now().AddDate(0, 0, -15),
    }
    budgetRepo.Create(budget)
    
    // Create transactions
    for i := 0; i < 5; i++ {
        tx := fixtures.NewTransaction().
            WithCategory(1).
            WithAmount(250.00).
            Build()
        txRepo.Create(tx)
    }
    
    // Check overspending
    overspent, amount := service.CheckOverspending(budget.ID)
    
    assert.True(t, overspent)
    assert.Equal(t, 250.00, amount) // $1250 spent - $1000 budget
}
```

### 3. Integration Tests
Test complete workflows:
```go
func TestCompleteTransactionWorkflow(t *testing.T) {
    db := test.SetupTestDB(t)
    
    // Initialize all components
    categoryRepo := repository.NewCategoryRepository(db)
    txRepo := repository.NewTransactionRepository(db)
    budgetRepo := repository.NewBudgetRepository(db)
    
    categoryService := service.NewCategoryService(categoryRepo)
    txService := service.NewTransactionService(txRepo, categoryService)
    budgetService := service.NewBudgetService(budgetRepo, txRepo)
    
    // Create category
    category := &models.Category{
        Name: "Groceries",
        Type: models.TransactionTypeExpense,
        Icon: "🛒",
    }
    err := categoryService.Create(category)
    require.NoError(t, err)
    
    // Create budget
    budget := &models.Budget{
        CategoryID: category.ID,
        Amount:     500.00,
        Period:     "monthly",
        StartDate:  time.Now().AddDate(0, 0, -1),
    }
    err = budgetService.Create(budget)
    require.NoError(t, err)
    
    // Create transaction
    tx := &models.Transaction{
        Type:        models.TransactionTypeExpense,
        Amount:      75.50,
        Currency:    "USD",
        CategoryID:  category.ID,
        Description: "Weekly groceries",
        Date:        time.Now(),
    }
    err = txService.Create(tx)
    require.NoError(t, err)
    
    // Verify budget status
    status, err := budgetService.GetStatus(budget.ID)
    require.NoError(t, err)
    
    assert.Equal(t, 75.50, status.Spent)
    assert.Equal(t, 424.50, status.Remaining)
    assert.Equal(t, 15.1, status.PercentUsed)
}
```

### 4. Currency Tests
Test exchange rate functionality:
```go
func TestCurrencyConversion_FixedRate(t *testing.T) {
    service := service.NewCurrencyService()
    
    // Test AED to USD (fixed rate)
    usdAmount, err := service.ConvertToUSD(100.00, "AED")
    require.NoError(t, err)
    assert.InDelta(t, 27.24, usdAmount, 0.01)
}

func TestCurrencyConversion_APIRate(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping API test in short mode")
    }
    
    service := service.NewCurrencyService()
    
    // Test EUR to USD (API rate)
    usdAmount, err := service.ConvertToUSD(100.00, "EUR")
    require.NoError(t, err)
    assert.Greater(t, usdAmount, 0.0)
}
```

## Test Helpers

### Data Cleanup
```go
func CleanupTestData(t *testing.T, db *gorm.DB) {
    t.Helper()
    
    db.Exec("DELETE FROM transactions")
    db.Exec("DELETE FROM categories")
    db.Exec("DELETE FROM budgets")
}
```

### Assertion Helpers
```go
func AssertTransactionEqual(t *testing.T, expected, actual *models.Transaction) {
    t.Helper()
    
    assert.Equal(t, expected.Type, actual.Type)
    assert.Equal(t, expected.Amount, actual.Amount)
    assert.Equal(t, expected.Currency, actual.Currency)
    assert.Equal(t, expected.CategoryID, actual.CategoryID)
    assert.Equal(t, expected.Description, actual.Description)
}
```

## Running Tests

### Commands
```bash
# Run all tests
make test

# Run tests with coverage
make test-cover

# Run tests without API calls
make test-short

# Run specific test
go test ./internal/repository -run TestTransactionCreate

# Run with verbose output
go test -v ./...
```

### Makefile Targets
```makefile
.PHONY: test test-short test-cover clean-test

test:
	@mkdir -p test/data
	@go test ./... -v

test-short:
	@mkdir -p test/data
	@go test -short ./... -v

test-cover:
	@mkdir -p test/data coverage
	@go test -coverprofile=coverage/coverage.out ./...
	@go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Coverage report: coverage/coverage.html"

clean-test:
	@rm -rf test/data/*
	@rm -rf coverage/*
```

## Coverage Goals

### Target Coverage
- Repository layer: 90%+
- Service layer: 85%+
- Models: 100%
- UI components: 70%+
- Overall: 80%+

### Coverage Exceptions
- Main function
- UI rendering code
- External API calls (tested with -short flag)
- Panic recovery code

## Best Practices

1. **Test Naming**: Use descriptive names that explain what is being tested
2. **Test Data**: Use fixtures for consistent test data
3. **Assertions**: Use testify for clear assertions
4. **Parallel Tests**: Mark independent tests as parallel
5. **Test Organization**: Group related tests in subtests

## Common Patterns

### Table-Driven Tests
```go
func TestTransactionValidation(t *testing.T) {
    tests := []struct {
        name    string
        tx      *models.Transaction
        wantErr bool
    }{
        {
            name: "valid transaction",
            tx: fixtures.NewTransaction().Build(),
            wantErr: false,
        },
        {
            name: "negative amount",
            tx: fixtures.NewTransaction().WithAmount(-100).Build(),
            wantErr: true,
        },
        {
            name: "invalid currency",
            tx: fixtures.NewTransaction().WithCurrency("XXX").Build(),
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.tx.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Benchmark Tests
```go
func BenchmarkTransactionRepository_GetAll(b *testing.B) {
    db := setupBenchDB(b)
    repo := repository.NewTransactionRepository(db)
    
    // Create 1000 transactions
    for i := 0; i < 1000; i++ {
        tx := fixtures.NewTransaction().Build()
        repo.Create(tx)
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, _ = repo.GetAll()
    }
}
```

## Continuous Integration

### GitHub Actions
```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.24'
    
    - name: Run tests
      run: make test-short
    
    - name: Generate coverage
      run: make test-cover
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage/coverage.out
```

## Debugging Failed Tests

1. **Check test output**: Use `-v` flag for verbose output
2. **Inspect database**: Keep test database with `KEEP_TEST_DB=1`
3. **Add logging**: Use `t.Logf()` for debug output
4. **Run single test**: Isolate failing test
5. **Check cleanup**: Ensure proper cleanup between tests

## Future Improvements

- [ ] Property-based testing for edge cases
- [ ] Mutation testing for test quality
- [ ] Performance regression tests
- [ ] Fuzz testing for input validation
- [ ] Contract testing for API integration