# BurnWise - Feature Documentation

## Project Focus: Monthly Burn Rate Tracking

BurnWise is primarily designed to help users monitor their monthly burn rate by providing clear visibility into recurring expenses and one-time transactions. The dashboard emphasizes expense tracking and projections based on recurring commitments.

## Completed Features

### 1. Transaction Management
- **CRUD Operations**: Create, read, update, and delete transactions
- **Multi-Currency Support**: Handle transactions in multiple currencies
- **Automatic Conversion**: Convert foreign currencies to USD for aggregation
- **Category Assignment**: Each transaction must have a category
- **Date Tracking**: Record transaction dates with time precision
- **Search & Filter**: Filter by date range, category, amount, or description

### 2. Category System
- **Default Categories**: Pre-configured income and expense categories
- **Icons & Colors**: Visual identification for each category
- **Type Separation**: Distinct categories for income vs expense
- **Usage Tracking**: Count transactions per category
- **Hierarchical Support**: Parent-child category relationships (model ready)

### 3. Budget Management
- **Monthly & Yearly Budgets**: Set budgets for different time periods
- **Category-Based**: Create budgets for specific spending categories
- **Real-Time Tracking**: Monitor spending against budget limits
- **Visual Progress**: Progress bars showing budget utilization
- **Overspending Alerts**: Color-coded warnings (green/yellow/red)
- **Period Calculations**: Automatic start/end date calculations

### 4. Financial Reports
- **Monthly Summary**: Income, expenses, and balance overview
- **Category Breakdown**: Spending by category with percentages
- **Time Navigation**: Browse reports by month/year
- **Visual Charts**: ASCII-based bar charts for categories
- **Budget vs Actual**: Compare spending against budgets

### 5. Data Export
- **CSV Export**: Export transactions, reports, and budgets
- **Command-Line Support**: Export via CLI flags
- **Flexible Output**: Write to file or stdout
- **Date Filtering**: Export specific time periods

### 6. Currency Configuration ✅ NEW
- **Enable/Disable Currencies**: Customize available currencies
- **Settings Persistence**: JSON-based configuration file
- **Validation**: Prevent disabling currencies with transactions
- **Fixed Exchange Rates**: Configure fixed rates (e.g., AED)
- **UI Management**: Interactive currency settings view
- **Default Currencies**: USD, EUR, AED enabled by default

### 7. User Interface
- **Keyboard-First**: Complete keyboard navigation
- **Vim-Style Navigation**: j/k for up/down movement
- **Modal Dialogs**: Forms for data entry
- **Real-Time Updates**: Instant feedback on actions
- **Help System**: Built-in keyboard shortcuts help
- **Responsive Layout**: Adapts to terminal size

### 7. Category Management ✅
- **Edit Categories**: Modify name, icon, and color of custom categories
- **Merge Categories**: Combine categories with atomic transaction migration
- **Create Custom**: Add new categories with custom properties
- **History Tracking**: Complete audit trail of all category changes
- **Delete Protection**: Prevent deletion of categories with transactions
- **Default Protection**: System categories cannot be modified
- **Type Safety**: Enforces category type consistency
- **UI Navigation**: Accessible via 'c' key from dashboard

### 8. Recurring Transactions ✅
- **Frequency Options**: Daily, weekly, monthly, yearly with custom intervals
- **Auto-Generation**: Creates transactions automatically when due
- **Manual Override**: Skip or modify individual occurrences
- **Pause/Resume**: Temporarily suspend recurring expenses
- **End Date Support**: Optional termination date
- **History Tracking**: Complete audit trail of skips/modifications
- **UI Management**: Accessible via 's' key from dashboard
- **Startup Processing**: Automatic processing of due transactions

## In-Progress Features

### Financial Projections (Phase 4)
**Status**: Algorithm designed, not implemented

- **Yearly Projections**: 12-month financial forecast
- **Recurring Integration**: Include recurring transactions
- **Historical Analysis**: Use past data for predictions
- **Confidence Levels**: Show projection reliability
- **Budget Comparison**: Project vs budget analysis

## Technical Features

### Data Management
- **SQLite Database**: Local storage with GORM ORM
- **Automatic Migrations**: Database schema updates
- **Transaction Safety**: ACID compliance
- **Soft Deletes**: Recoverable deletion
- **Indexes**: Performance optimization

### Architecture
- **Repository Pattern**: Clean data access layer
- **Service Layer**: Business logic separation
- **Dependency Injection**: Testable design
- **Error Handling**: Graceful error management
- **Concurrent Safety**: Thread-safe operations

### Testing
- **Real Database Tests**: No mocks, actual SQLite
- **Automatic Cleanup**: Test isolation
- **Comprehensive Coverage**: Unit and integration tests
- **Test Helpers**: Reusable test utilities
- **Skip Flags**: Offline test support

### Performance
- **Lazy Loading**: Efficient data retrieval
- **Caching**: Exchange rate caching (1 hour)
- **Pagination**: Handle large datasets
- **Query Optimization**: Efficient SQL queries
- **Minimal Dependencies**: Fast startup

## Configuration & Settings

### Settings File Structure
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
    "decimal_places": 2,
    "theme": "default"
  },
  "version": "1.0.0"
}
```

### Supported Currencies
38+ currencies including:
- Major: USD, EUR, GBP, JPY, CHF, CAD, AUD
- Asian: CNY, INR, KRW, SGD, HKD, THB, MYR
- Americas: BRL, MXN, ARS, CLP, COP
- Others: ZAR, TRY, NOK, SEK, DKK

## Keyboard Shortcuts

### Global
- `q` - Quit application
- `Esc` - Cancel/Back
- `Tab` - Next field
- `Shift+Tab` - Previous field
- `/` - Quick search

### Dashboard
- `n` - New transaction
- `t` - View transactions
- `b` - Manage budgets
- `r` - View reports
- `c` - Manage categories
- `s` - Manage recurring expenses
- `u` - Currency settings

### Lists
- `j`/`↓` - Move down
- `k`/`↑` - Move up
- `Enter` - Select/Edit
- `d` - Delete with confirmation
- `f` - Filter options

### Forms
- `Tab` - Navigate fields
- `Enter` - Save
- `Esc` - Cancel without saving

### Category Management
- `n` - Create new category
- `e` - Edit selected category
- `m` - Merge categories
- `d` - Delete empty category
- `Enter` - Select category
- `Esc` - Back to dashboard

## API Integration

### Exchange Rate API
- **Provider**: ExchangeRate-API (v4)
- **Free Tier**: 1000 requests/month
- **Cache Duration**: 1 hour
- **Base Currency**: USD
- **Fallback**: Fixed rates for configured currencies

## File Locations

### Database
- Linux/Mac: `~/.local/share/burnwise/burnwise.db`
- Fallback: `./data/burnwise.db`

### Settings
- Location: `./data/settings.json`
- Auto-created on first run

### Logs
- Location: `./data/burnwise.log`
- Debug mode: `BUDGET_DEBUG=1`

## Security Considerations

- Local data storage only
- No network requests except exchange rates
- No authentication (single-user)
- File permissions respected
- No sensitive data logging

## Future Enhancements

### Planned
- Email notifications for budget alerts
- Data backup/restore commands
- Receipt attachments
- Financial goals tracking
- Bill reminders
- Multi-account support

### Under Consideration
- Cloud sync option
- Mobile companion app
- Investment tracking
- Tax preparation tools
- Financial insights AI
- Multi-user support