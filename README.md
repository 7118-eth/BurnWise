# Budget Tracker

A fast, keyboard-driven terminal application for personal finance management built with Go.

## Features

- 💰 **Income & Expense Tracking** - Record all your financial transactions
- 🌍 **Multi-Currency Support** - Track expenses in multiple currencies with automatic conversion
- 🔧 **Configurable Currencies** - Enable/disable currencies based on your needs
- 📊 **Budget Management** - Set monthly/yearly budgets and track progress
- 📈 **Financial Reports** - View spending trends and category breakdowns
- ⌨️ **Keyboard-First Design** - Navigate entirely with keyboard shortcuts
- 🔍 **Smart Search** - Filter transactions by date, category, or amount
- 📁 **Data Export** - Export your data to CSV for external analysis
- ⚙️ **Settings Management** - JSON-based configuration with currency preferences

## Installation

### Prerequisites
- Go 1.24 or higher
- Make (optional, for using Makefile commands)

### From Source
```bash
git clone https://github.com/yourusername/budget-tracker
cd budget-tracker
make build
```

### Install
```bash
make install  # Installs to $GOPATH/bin
```

## Quick Start

1. **Launch the application**
   ```bash
   budget
   ```

2. **Add your first transaction** - Press `n` to create a new transaction

3. **Set up budgets** - Press `b` to manage your monthly budgets

4. **View reports** - Press `r` to see your financial summary

## Usage

### Main Dashboard
```
💰 Budget Tracker                                    October 2025

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

[n]ew  [t]ransactions  [b]udgets  [r]eports  c[u]rrencies  [q]uit
```

### Keyboard Shortcuts

#### Global
- `q` - Quit application
- `Esc` - Cancel/go back
- `/` - Quick search
- `?` - Show help

#### Navigation
- `j` or `↓` - Move down
- `k` or `↑` - Move up
- `Tab` - Next field
- `Shift+Tab` - Previous field
- `Enter` - Select/confirm

#### Actions
- `n` - New transaction
- `t` - View all transactions
- `b` - Manage budgets
- `r` - View reports
- `u` - Currency settings
- `c` - Manage categories (coming soon)
- `e` - Edit selected item
- `d` - Delete selected item (with confirmation)
- `f` - Filter options

### Adding Transactions

1. Press `n` from the main screen
2. Fill in the transaction details:
   - Type: Income or Expense
   - Amount: Enter the value
   - Currency: Select from dropdown
   - Category: Choose appropriate category
   - Description: Brief note about the transaction
   - Date: Defaults to today, can be changed

### Managing Budgets

1. Press `b` from the main screen
2. Press `n` to create a new budget
3. Select a category and set monthly limit
4. Track spending against budgets in real-time

### Currency Management

Press `u` from the dashboard to access currency settings where you can:
- Enable/disable currencies for your transactions
- View which currencies are currently active
- See fixed exchange rates (e.g., AED: 1 USD = 3.6725 AED)

Default enabled currencies:
- **USD** - US Dollar (base currency)
- **EUR** - Euro
- **AED** - UAE Dirham (fixed rate)

You can enable additional currencies from a list of 38+ supported currencies including GBP, JPY, CHF, CAD, AUD, CNY, INR, and more.

## Data Storage

Your financial data is stored locally:
- **Database**: 
  - Linux/Mac: `~/.local/share/budget-tracker/budget.db`
  - Windows: `%APPDATA%\budget-tracker\budget.db`
- **Settings**: `./data/settings.json`

### Backup

To backup your data:
```bash
budget export --format csv --output backup.csv
```

To backup the entire database:
```bash
cp ~/.local/share/budget-tracker/budget.db budget-backup.db
```

## Configuration

The application uses a JSON settings file (`data/settings.json`) that is automatically created on first run:

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

### Settings Explained

- **currencies.enabled**: List of currencies available in the application
- **currencies.default**: Default currency for new transactions
- **currencies.fixed_rates**: Currencies with fixed exchange rates (not fetched from API)
- **ui.date_format**: Date display format (Go time format)
- **ui.decimal_places**: Number of decimal places for amounts
- **ui.theme**: UI theme (currently only "default")

## Development

### Building from Source
```bash
# Clone the repository
git clone https://github.com/yourusername/budget-tracker
cd budget-tracker

# Install dependencies
go mod download

# Build the application
make build

# Run tests
make test

# Run with live reload during development
make dev
```

### Project Structure
```
budget-tracker/
├── cmd/budget/         # Application entry point
├── internal/
│   ├── models/        # Data models
│   ├── repository/    # Database access
│   ├── service/       # Business logic
│   ├── ui/           # Terminal UI
│   └── db/           # Database setup
├── test/             # Test utilities
└── data/             # SQLite database
```

### Running Tests
```bash
make test          # Run all tests
make test-short    # Skip integration tests
make test-cover    # Generate coverage report
```

## Troubleshooting

### Application won't start
- Check if port is already in use
- Verify Go version: `go version` (requires 1.24+)
- Check database permissions

### Exchange rates not updating
- Verify internet connection
- Check API rate limits (1000 requests/month on free tier)
- Rates are cached for 1 hour

### Database errors
- Ensure write permissions in data directory
- Check disk space
- Run database integrity check: `budget check-db`

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`make test`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [GORM](https://gorm.io/) - Database ORM
- [ExchangeRate-API](https://exchangerate-api.com/) - Currency conversion
- [SQLite](https://sqlite.org/) - Database engine