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
