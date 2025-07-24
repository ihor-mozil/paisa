package server

import (
	"sort"
	"time"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func GetBudgetCustom(db *gorm.DB) gin.H {
	forecastPostings := query.Init(db).Like("Expenses:%").Forecast().All()
	expenses := query.Init(db).Like("Expenses:%").All()
	return computeBudetCustom(db, forecastPostings, expenses)
}

func GetCurrentBudgetCustom(db *gorm.DB) gin.H {
	forecastPostings := query.Init(db).Like("Expenses:%").Forecast().UntilThisMonthEnd().All()
	expenses := query.Init(db).Like("Expenses:%").UntilThisMonthEnd().All()
	return computeBudetCustom(db, forecastPostings, expenses)
}

func computeBudetCustom(db *gorm.DB, forecastPostings, expensesPostings []posting.Posting) gin.H {
	postings := query.Init(db).AccountPrefix("Assets:Checking").All()
	postings = service.PopulateMarketPrice(db, postings)
	checkingBalance := accounting.CurrentBalance(postings)
	availableForBudgeting := checkingBalance

	forecasts := utils.GroupByMonth(forecastPostings)
	expenses := utils.GroupByMonth(expensesPostings)

	accounts := lo.Uniq(lo.Map(forecastPostings, func(p posting.Posting, _ int) string {
		return p.Account
	}))
	sort.Strings(accounts)

	budgetsByMonth := make(map[string]Budget)
	balance := make(map[string]decimal.Decimal)

	currentMonth := lo.Must(time.ParseInLocation("2006-01", utils.Now().Format("2006-01"), config.TimeZone()))

	if len(forecastPostings) > 0 {
		start := utils.BeginningOfMonth(forecastPostings[0].Date)
		end := utils.EndOfMonth(forecastPostings[len(forecastPostings)-1].Date)

		for start := start; start.Before(end) || start.Equal(end); start = start.AddDate(0, 1, 0) {
			month := start.Format("2006-01")
			var accountBudgets []AccountBudget

			forecastsByMonth := forecasts[month]
			date := lo.Must(time.ParseInLocation("2006-01", month, config.TimeZone()))
			expensesByMonth, ok := expenses[month]
			if !ok {
				expensesByMonth = []posting.Posting{}
			}

			forecastsByAccount := accounting.GroupByAccount(forecastsByMonth)
			expensesByAccount := accounting.GroupByAccount(expensesByMonth)

			for _, account := range accounts {
				fs := forecastsByAccount[account]
				es := popExpenses(account, expensesByAccount)
				if !ok {
					es = []posting.Posting{}
				}

				budget := buildBudget(date, account, balance[account], fs, es, date.Before(currentMonth))
				// if budget.Available.IsPositive() {
				balance[account] = budget.Available
				// } else {
				// 	balance[account] = decimal.Zero
				// }

				accountBudgets = append(accountBudgets, budget)
			}

			availableThisMonth := utils.SumBy(
				accountBudgets, func(budget AccountBudget) decimal.Decimal {
					// if budget.Available.IsPositive() {
					return budget.Available
					// }
					// return decimal.Zero
				})

			forecast := utils.SumBy(
				accountBudgets, func(budget AccountBudget) decimal.Decimal {
					if budget.Forecast.IsPositive() {
						return budget.Forecast
					}
					return decimal.Zero
				})

			availableForBudgeting = availableThisMonth
			endOfMonthBalance := availableThisMonth.Add(checkingBalance)

			budgetsByMonth[month] = Budget{
				Date:               date,
				Accounts:           accountBudgets,
				EndOfMonthBalance:  endOfMonthBalance,
				AvailableThisMonth: availableThisMonth,
				Forecast:           forecast,
			}
		}
	}

	return gin.H{
		"budgetsByMonth":        budgetsByMonth,
		"checkingBalance":       checkingBalance,
		"availableForBudgeting": availableForBudgeting,
	}
}
