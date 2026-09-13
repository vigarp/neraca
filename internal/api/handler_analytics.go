package api

import (
	"database/sql"
	"math"
	"net/http"
	"strconv"
	"time"

	"neraca/internal/database"
	"neraca/internal/models"
)

func getPaydayInMonth(year int, month time.Month, targetDay int) time.Time {
	// Hari terakhir dalam bulan tersebut diperoleh dari hari ke-0 bulan berikutnya
	lastDayOfMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
	actualDay := targetDay
	if actualDay > lastDayOfMonth {
		actualDay = lastDayOfMonth
	}
	return time.Date(year, month, actualDay, 0, 0, 0, 0, time.Local)
}

func calculatePaydayCycle(now time.Time, paydayDay int) (time.Time, time.Time, int, int) {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	thisMonthPayday := getPaydayInMonth(today.Year(), today.Month(), paydayDay)

	var cycleStart, nextPayday time.Time
	if today.Before(thisMonthPayday) {
		nextPayday = thisMonthPayday
		cycleStart = getPaydayInMonth(today.Year(), today.Month()-1, paydayDay)
	} else {
		cycleStart = thisMonthPayday
		nextPayday = getPaydayInMonth(today.Year(), today.Month()+1, paydayDay)
	}

	diffRemain := int(nextPayday.Sub(today).Hours() / 24)
	daysRemain := diffRemain
	if daysRemain <= 0 {
		daysRemain = 1
	}

	diffElapsed := int(today.Sub(cycleStart).Hours() / 24)
	daysElapsed := diffElapsed
	if daysElapsed <= 0 {
		daysElapsed = 1
	}

	return cycleStart, nextPayday, daysRemain, daysElapsed
}

func fetchPaydaySetting(db *database.DB) int {
	var val string
	err := db.QueryRow("SELECT value FROM settings WHERE key = 'payday_date'").Scan(&val)
	if err == nil {
		if d, err := strconv.Atoi(val); err == nil && d >= 1 && d <= 31 {
			return d
		}
	}
	return 25 // default tanggal 25
}

func fetchOperationalBalance(db *database.DB) (float64, error) {
	var total float64
	query := `
	SELECT COALESCE(SUM(balance), 0)
	FROM accounts
	WHERE account_group = 'operational' AND is_active = 1;
	`
	err := db.QueryRow(query).Scan(&total)
	return total, err
}

func fetchCycleExpenses(
	db *database.DB,
	startDate, endDate string,
) (float64, models.PillarBreakdown, []models.CategoryExpenseBreakdown, error) {
	query := `
	SELECT 
		COALESCE(c.id, 0) AS category_id,
		COALESCE(c.name, 'Lainnya') AS category_name,
		COALESCE(c.pillar, 'needs') AS pillar,
		COALESCE(c.icon, '') AS icon,
		COALESCE(c.color, '') AS color,
		SUM(t.amount) AS total,
		COUNT(t.id) AS transaction_count
	FROM transactions t
	LEFT JOIN categories c ON t.category_id = c.id
	WHERE t.type = 'expense'
	  AND t.transaction_date >= ?
	  AND t.transaction_date < ?
	GROUP BY c.id, c.name, c.pillar, c.icon, c.color
	ORDER BY total DESC;
	`
	rows, err := db.Query(query, startDate, endDate)
	if err != nil {
		return 0, models.PillarBreakdown{}, nil, err
	}
	defer rows.Close()

	var grandTotal float64
	categories := []models.CategoryExpenseBreakdown{}
	pillarTotals := map[string]float64{
		"needs":   0,
		"wants":   0,
		"savings": 0,
	}

	for rows.Next() {
		var item models.CategoryExpenseBreakdown
		if err := rows.Scan(
			&item.ID, &item.Name, &item.Pillar, &item.Icon, &item.Color, &item.Total, &item.TransactionCount,
		); err != nil {
			return 0, models.PillarBreakdown{}, nil, err
		}

		grandTotal += item.Total
		pillarTotals[item.Pillar] += item.Total
		categories = append(categories, item)
	}

	// Hitung persentase
	for i := range categories {
		if grandTotal > 0 {
			categories[i].Percentage = math.Round((categories[i].Total/grandTotal)*1000) / 10
		}
	}

	pillarBreakdown := models.PillarBreakdown{
		Needs:   models.PillarExpenseStat{Total: pillarTotals["needs"]},
		Wants:   models.PillarExpenseStat{Total: pillarTotals["wants"]},
		Savings: models.PillarExpenseStat{Total: pillarTotals["savings"]},
	}

	if grandTotal > 0 {
		pillarBreakdown.Needs.Percentage = math.Round((pillarTotals["needs"]/grandTotal)*1000) / 10
		pillarBreakdown.Wants.Percentage = math.Round((pillarTotals["wants"]/grandTotal)*1000) / 10
		pillarBreakdown.Savings.Percentage = math.Round((pillarTotals["savings"]/grandTotal)*1000) / 10
	}

	return grandTotal, pillarBreakdown, categories, nil
}

func fetchCycleIncome(db *database.DB, startDate, endDate string) (float64, error) {
	var total float64
	query := `
	SELECT COALESCE(SUM(amount), 0)
	FROM transactions
	WHERE type = 'income'
	  AND transaction_date >= ?
	  AND transaction_date < ?;
	`
	err := db.QueryRow(query, startDate, endDate).Scan(&total)
	return total, err
}

func determineBurnRateStatus(avgExpense, dailyLimit float64) string {
	if dailyLimit <= 0 || avgExpense <= dailyLimit {
		return "safe"
	}
	if avgExpense <= dailyLimit*1.2 {
		return "warning"
	}
	return "danger"
}

func parseTargetDate(r *http.Request) time.Time {
	if dateParam := r.URL.Query().Get("date"); dateParam != "" {
		if t, err := time.Parse(time.DateOnly, dateParam); err == nil {
			return t
		}
	}
	return time.Now()
}

func calculateDailyLimit(remainFunds float64, daysRemain int) float64 {
	if remainFunds <= 0 || daysRemain <= 0 {
		return 0
	}
	return math.Round((remainFunds/float64(daysRemain))*100) / 100
}

func calculateAverageExpense(grandTotal float64, daysElapsed int) float64 {
	if grandTotal <= 0 || daysElapsed <= 0 {
		return 0
	}
	return math.Round((grandTotal/float64(daysElapsed))*100) / 100
}

func handleGetBurnRateAnalytics(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := parseTargetDate(r)
		paydayDay := fetchPaydaySetting(db)
		cycleStart, nextPayday, daysRemain, daysElapsed := calculatePaydayCycle(now, paydayDay)

		remainFunds, err := fetchOperationalBalance(db)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		startDateStr := cycleStart.Format(time.DateOnly)
		endDateStr := nextPayday.Format(time.DateOnly)

		grandTotal, pillarBreakdown, categories, err := fetchCycleExpenses(db, startDateStr, endDateStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cycleIncome, err := fetchCycleIncome(db, startDateStr, endDateStr)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		prospectDailyLimit := calculateDailyLimit(remainFunds, daysRemain)
		avgDailyExpense := calculateAverageExpense(grandTotal, daysElapsed)
		status := determineBurnRateStatus(avgDailyExpense, prospectDailyLimit)

		res := models.BurnRateAnalyticsResponse{
			NextPaydayRemain:    daysRemain,
			NextPaydayDate:      endDateStr,
			CycleStartDate:      startDateStr,
			DaysElapsed:         daysElapsed,
			RemainFunds:         remainFunds,
			GrandTotalExpenses:  grandTotal,
			ProspectDailyLimit:  prospectDailyLimit,
			AverageDailyExpense: avgDailyExpense,
			BurnRateStatus:      status,
			CycleIncome:         cycleIncome,
			PillarBreakdown:     pillarBreakdown,
			CategoryBreakdown:   categories,
		}

		writeJSON(w, http.StatusOK, res)
	}
}
