package models

import "time"

type Account struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	AccountGroup string    `json:"account_group"` // operational, passive_asset, emergency
	Type         string    `json:"type"`          // bank, ewallet, cash, mutual_fund, gold, stocks, crypto, other
	Balance      float64   `json:"balance"`
	Currency     string    `json:"currency"`
	Institution  string    `json:"institution"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateAccountRequest struct {
	Name         string  `json:"name"`
	AccountGroup string  `json:"account_group"`
	Type         string  `json:"type"`
	Balance      float64 `json:"balance"`
	Currency     string  `json:"currency"`
	Institution  string  `json:"institution"`
}

type UpdateAccountRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Institution string `json:"institution"`
}

type RevalueAccountRequest struct {
	NewBalance float64 `json:"new_balance"`
	Notes      string  `json:"notes"`
}

type AssetValuation struct {
	ID              int64     `json:"id"`
	AccountID       int64     `json:"account_id"`
	PreviousBalance float64   `json:"previous_balance"`
	NewBalance      float64   `json:"new_balance"`
	Difference      float64   `json:"difference"`
	Notes           string    `json:"notes"`
	RecordedAt      time.Time `json:"recorded_at"`
}

type AccountsSummaryResponse struct {
	OperationalAccounts     []Account `json:"operational_accounts"`
	PassiveAccounts         []Account `json:"passive_accounts"`
	EmergencyAccounts       []Account `json:"emergency_accounts"`
	TotalOperationalBalance float64   `json:"total_operational_balance"`
	TotalPassiveAssets      float64   `json:"total_passive_assets"`
	TotalEmergencyBalance   float64   `json:"total_emergency_balance"`
	TotalNetWorth           float64   `json:"total_net_worth"`
}

type SettingsResponse struct {
	PaydayDate          int     `json:"payday_date"`
	MonthlyIncomeBudget float64 `json:"monthly_income_budget"`
}

type UpdateSettingsRequest struct {
	PaydayDate          int     `json:"payday_date"`
	MonthlyIncomeBudget float64 `json:"monthly_income_budget"`
}

type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`   // 'income', 'expense'
	Pillar    string    `json:"pillar"` // 'needs', 'wants', 'savings', 'income'
	Icon      string    `json:"icon"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateCategoryRequest struct {
	Name   string `json:"name"`
	Type   string `json:"type"`   // 'income', 'expense'
	Pillar string `json:"pillar"` // 'needs', 'wants', 'savings', 'income'
	Icon   string `json:"icon"`
	Color  string `json:"color"`
}

type UpdateCategoryRequest struct {
	Name   string `json:"name"`
	Pillar string `json:"pillar"`
	Icon   string `json:"icon"`
	Color  string `json:"color"`
}

type CategoriesSummaryResponse struct {
	NeedsCategories   []Category `json:"needs_categories"`
	WantsCategories   []Category `json:"wants_categories"`
	SavingsCategories []Category `json:"savings_categories"`
	IncomeCategories  []Category `json:"income_categories"`
	AllCategories     []Category `json:"all_categories"`
}

type Transaction struct {
	ID              int64     `json:"id"`
	AccountID       int64     `json:"account_id"`
	AccountName     string    `json:"account_name"`
	ToAccountID     *int64    `json:"to_account_id,omitempty"`
	ToAccountName   string    `json:"to_account_name,omitempty"`
	CategoryID      *int64    `json:"category_id,omitempty"`
	CategoryName    string    `json:"category_name,omitempty"`
	Pillar          string    `json:"pillar,omitempty"` // 'needs', 'wants', 'savings', 'income'
	Amount          float64   `json:"amount"`
	Type            string    `json:"type"` // 'expense', 'income', 'transfer'
	Description     string    `json:"description"`
	TransactionDate string    `json:"transaction_date"` // 'YYYY-MM-DD'
	CreatedAt       time.Time `json:"created_at"`
}

type CreateTransactionRequest struct {
	AccountID       int64   `json:"account_id"`
	CategoryID      *int64  `json:"category_id"`
	Amount          float64 `json:"amount"`
	Type            string  `json:"type"` // 'expense', 'income'
	Description     string  `json:"description"`
	TransactionDate string  `json:"transaction_date"` // 'YYYY-MM-DD'
}

type CreateTransferRequest struct {
	FromAccountID   int64   `json:"from_account_id"`
	ToAccountID     int64   `json:"to_account_id"`
	CategoryID      *int64  `json:"category_id,omitempty"`
	Amount          float64 `json:"amount"`
	Description     string  `json:"description"`
	TransactionDate string  `json:"transaction_date"` // 'YYYY-MM-DD'
}

type TransactionsResponse struct {
	Transactions []Transaction `json:"transactions"`
	TotalExpense float64       `json:"total_expense"`
	TotalIncome  float64       `json:"total_income"`
}

type PillarExpenseStat struct {
	Total      float64 `json:"total"`
	Percentage float64 `json:"percentage"`
}

type PillarBreakdown struct {
	Needs   PillarExpenseStat `json:"needs"`
	Wants   PillarExpenseStat `json:"wants"`
	Savings PillarExpenseStat `json:"savings"`
}

type CategoryExpenseBreakdown struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	Pillar           string  `json:"pillar"`
	Icon             string  `json:"icon"`
	Color            string  `json:"color"`
	Total            float64 `json:"total"`
	Percentage       float64 `json:"percentage"`
	TransactionCount int     `json:"transaction_count"`
}

type BurnRateAnalyticsResponse struct {
	NextPaydayRemain    int                        `json:"next_payday_remain"`
	NextPaydayDate      string                     `json:"next_payday_date"`
	CycleStartDate      string                     `json:"cycle_start_date"`
	DaysElapsed         int                        `json:"days_elapsed"`
	RemainFunds         float64                    `json:"remain_funds"`
	GrandTotalExpenses  float64                    `json:"grand_total_expenses"`
	ProspectDailyLimit  float64                    `json:"prospect_daily_limit"`
	AverageDailyExpense float64                    `json:"average_daily_expense"`
	BurnRateStatus      string                     `json:"burn_rate_status"` // 'safe', 'warning', 'danger'
	CycleIncome         float64                    `json:"cycle_income"`
	PillarBreakdown     PillarBreakdown            `json:"pillar_breakdown"`
	CategoryBreakdown   []CategoryExpenseBreakdown `json:"category_breakdown"`
}
