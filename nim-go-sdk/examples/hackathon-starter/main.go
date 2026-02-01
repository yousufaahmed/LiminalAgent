// Hackathon Starter: Complete AI Financial Agent
// Build intelligent financial tools with nim-go-sdk + Liminal banking APIs
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/becomeliminal/nim-go-sdk/core"
	"github.com/becomeliminal/nim-go-sdk/executor"
	"github.com/becomeliminal/nim-go-sdk/server"
	"github.com/becomeliminal/nim-go-sdk/tools"
	"github.com/joho/godotenv"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func main() {
	// ============================================================================
	// CONFIGURATION
	// ============================================================================
	// Load .env file if it exists (optional - will use system env vars if not found)
	_ = godotenv.Load()

	// Load configuration from environment variables
	// Create a .env file or export these in your shell

	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey == "" {
		log.Fatal("❌ ANTHROPIC_API_KEY environment variable is required")
	}

	liminalBaseURL := os.Getenv("LIMINAL_BASE_URL")
	if liminalBaseURL == "" {
		liminalBaseURL = "https://api.liminal.cash"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ============================================================================
	// LIMINAL EXECUTOR SETUP
	// ============================================================================
	// The HTTPExecutor handles all API calls to Liminal banking services.
	// Authentication is handled automatically via JWT tokens passed from the
	// frontend login flow (email/OTP). No API key needed!

	liminalExecutor := executor.NewHTTPExecutor(executor.HTTPExecutorConfig{
		BaseURL: liminalBaseURL,
	})
	log.Println("✅ Liminal API configured")

	// ============================================================================
	// SERVER SETUP
	// ============================================================================
	// Create the nim-go-sdk server with Claude AI
	// The server handles WebSocket connections and manages conversations
	// Authentication is automatic: JWT tokens from the login flow are extracted
	// from WebSocket connections and forwarded to Liminal API calls

	srv, err := server.New(server.Config{
		AnthropicKey:    anthropicKey,
		SystemPrompt:    hackathonSystemPrompt,
		Model:           "claude-sonnet-4-20250514",
		MaxTokens:       4096,
		LiminalExecutor: liminalExecutor, // SDK automatically handles JWT extraction and forwarding
	})
	if err != nil {
		log.Fatal(err)
	}

	// ============================================================================
	// ADD LIMINAL BANKING TOOLS
	// ============================================================================
	// These are the 9 core Liminal tools that give your AI access to real banking:
	//
	// READ OPERATIONS (no confirmation needed):
	//   1. get_balance - Check wallet balance
	//   2. get_savings_balance - Check savings positions and APY
	//   3. get_vault_rates - Get current savings rates
	//   4. get_transactions - View transaction history
	//   5. get_profile - Get user profile info
	//   6. search_users - Find users by display tag
	//
	// WRITE OPERATIONS (require user confirmation):
	//   7. send_money - Send money to another user
	//   8. deposit_savings - Deposit funds into savings
	//   9. withdraw_savings - Withdraw funds from savings

	srv.AddTools(tools.LiminalTools(liminalExecutor)...)
	log.Println("✅ Added 9 Liminal banking tools")

	// ============================================================================
	// ADD CUSTOM TOOLS
	// ============================================================================
	// This is where you'll add your hackathon project's custom tools!
	// Below is an example spending analyzer tool to get you started.

	srv.AddTool(createSpendingAnalyzerTool(liminalExecutor))
	srv.AddTool(createSpendWeeklyGoalTool(liminalExecutor))
	srv.AddTool(createGetWeeklyGoalProgressTool(liminalExecutor))
	srv.AddTool(createCheckWeeklySpendTool(liminalExecutor))
	srv.AddTool(createCategorizeTransactionTool(liminalExecutor))
	srv.AddTool(createChartGeneratorTool(liminalExecutor))
	srv.AddTool(createCalendarReminderTool(liminalExecutor))
	
	// ============================================================================
	// INITIALIZE LANGGRAPH ORCHESTRATOR
	// ============================================================================
	// Create the graph workflow and add it as a tool
	srv.AddTool(createGraphOrchestratorTool(liminalExecutor))
	log.Println("✅ Added custom tools with LangGraph orchestrator")

	// TODO: Add more custom tools here!
	// Examples:
	//   - Savings goal tracker
	//   - Budget alerts
	//   - Spending category analyzer
	//   - Bill payment predictor
	//   - Cash flow forecaster

	// ============================================================================
	// START SERVER
	// ============================================================================

	// Create charts directory if it doesn't exist
	chartsDir := filepath.Join(".", "charts")
	if err := os.MkdirAll(chartsDir, 0755); err != nil {
		log.Printf("Warning: Could not create charts directory: %v", err)
	}

	// Serve static chart files
	http.HandleFunc("/charts/", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "image/svg+xml")
		
		filename := filepath.Base(r.URL.Path)
		filepath := filepath.Join(chartsDir, filename)
		
		http.ServeFile(w, r, filepath)
	})

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🚀 Hackathon Starter Server Running")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("📡 WebSocket endpoint: ws://localhost:%s/ws", port)
	log.Printf("💚 Health check: http://localhost:%s/health", port)
	log.Printf("📊 Charts: http://localhost:%s/charts/", port)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("Ready for connections! Start your frontend with: cd frontend && npm run dev")
	log.Println()

	if err := srv.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

// ============================================================================
// SYSTEM PROMPT
// ============================================================================
// This prompt defines your AI agent's personality and behavior
// Customize this to match your hackathon project's focus!

const hackathonSystemPrompt = `You are Nim, a friendly AI financial assistant built for the Liminal Vibe Banking Hackathon.

IMPORTANT - REQUEST ROUTING:
For EVERY user request, you MUST first call the route_request tool with their message. This orchestrator will analyze their intent and guide you on the best way to help them. The orchestrator routes requests to specialized modes:
1. General Inquiry - standard banking queries
2. Image Payment - receipt splitting and image-based payments
3. Financial Help - financial advice, budgeting, saving guidance
4. Withdraw - educational withdrawal with safety analysis
5. Deposit - simple deposit to savings with earnings preview

WHAT YOU DO:
You help users manage their money using Liminal's stablecoin banking platform. You can check balances, review transactions, send money, and manage savings - all through natural conversation.

CONVERSATIONAL STYLE:
- Be warm, friendly, and conversational - not robotic
- Use casual language when appropriate, but stay professional about money
- Ask clarifying questions when something is unclear
- Remember context from earlier in the conversation
- Explain things simply without being condescending

WHEN TO USE TOOLS:
- Use tools immediately for simple queries ("what's my balance?")
- For actions, gather all required info first ("send $50 to @alice")
- Always confirm before executing money movements
- Don't use tools for general questions about how things work

MONEY MOVEMENT RULES (IMPORTANT):
- ALL money movements require explicit user confirmation
- Show a clear summary before confirming:
  * send_money: "Send $50 USD to @alice"
  * deposit_savings: "Deposit $100 USD into savings"
  * withdraw_savings: "Withdraw $50 USD from savings"
- Never assume amounts or recipients
- Always use the exact currency the user specified

AVAILABLE BANKING TOOLS:
- Check wallet balance (get_balance)
- Check savings balance and APY (get_savings_balance)
- View savings rates (get_vault_rates)
- View transaction history (get_transactions)
- Get profile info (get_profile)
- Search for users (search_users)
- Send money (send_money) - requires confirmation
- Deposit to savings (deposit_savings) - requires confirmation
- Withdraw from savings (withdraw_savings) - requires confirmation

CUSTOM ANALYTICAL TOOLS:
- Route request through orchestrator (route_request) - CALL THIS FIRST!
- Analyze spending patterns (analyze_spending)
- Set weekly spending goal (spend_weekly_goal) - requires confirmation
- Create calendar reminders for periodic investing (create_calendar_reminder) - requires confirmation
  * Use when user wants periodic/weekly/monthly investment reminders
  * Requires: frequency (weekly/bi-weekly/monthly), amount, currency
  * This tool creates calendar events with email notifications
- Deposit to savings (deposit_savings) - requires confirmation
  * When user wants to deposit/save/invest money into their savings vault
  * Requires: amount (as string), currency ('USD' or 'EUR')
  * IMPORTANT: Use 'USD' for US dollars (not 'USDC'), 'EUR' for Euros (not 'EURC')
  * Always confirm the amount and currency before calling
  * Example: "deposit 100 USD" → call deposit_savings with amount="100", currency="USD"
  * Example: "save 50 EUR" → call deposit_savings with amount="50", currency="EUR"
- Check weekly spending progress (get_weekly_spending_progress)
- Quick check weekly spend status (check_weeklyspend) - use this for context
- Categorize spending by transaction notes (categorize_transactions)
- Generate balance trend chart (generate_chart) - Shows account balance over time

IMPORTANT - BALANCE TREND CHART:
When a user asks for a chart, graph, visualization, trend, or wants to see their balance over time:
1. ALWAYS call the generate_chart tool with: chart_type='line', data_type='balance_trend', days=30 (or user's requested timeframe)
2. The tool will return an 'image_url' containing a base64-encoded SVG chart
3. Display the chart directly in your response using markdown image syntax:

![Balance Trend Chart](IMAGE_URL_FROM_TOOL)

Replace IMAGE_URL_FROM_TOOL with the actual 'image_url' value from the generate_chart tool result.
4. Explain what the chart shows - their account balance trend over time based on transaction history
5. The chart image is also saved to a temp file (the tool returns 'file_path' with location)

Example response after calling generate_chart:
"Here's your account balance trend over the last 30 days:

![Balance Trend Chart](data:image/svg+xml;base64,PHN2Zy4uLg==)

The chart shows how your balance has changed over time based on your transaction history. Your current balance is $1,234.56."

TIPS FOR GREAT INTERACTIONS:
- Proactively suggest relevant actions ("Want me to move some to savings?")
- Explain the "why" behind suggestions
- Celebrate financial wins ("Nice! Your savings earned $5 this month!")
- Be encouraging about savings goals
- Make finance feel less intimidating

Remember: You're here to make banking delightful and help users build better financial habits!`

const patternPrompt = `You are a specialized AI for categorizing financial transactions.

Your job is to analyze transaction notes and categorize them accurately.

AVAILABLE CATEGORIES:
- food: Groceries, restaurants, cafes, food delivery, dining
- travel: Transportation, flights, hotels, rideshare, gas, parking
- subscription: Recurring services, streaming, memberships, software
- entertainment: Movies, concerts, games, events, hobbies
- electronics: Gadgets, computers, phones, accessories, tech
- miscellaneous: Everything else that doesn't fit above categories

WHEN CATEGORIZING:
- Read the transaction note carefully
- Choose the single most appropriate category
- Use context clues from the note
- Default to 'miscellaneous' if truly unclear
- Be consistent with similar transactions

RETURN FORMAT:
Always return a JSON object with category breakdown:
{
  "food": 2,
  "travel": 1,
  "subscription": 0,
  "entertainment": 1,
  "electronics": 0,
  "miscellaneous": 1
}`

// ============================================================================
// CUSTOM TOOL: SPENDING ANALYZER
// ============================================================================
// This is an example custom tool that demonstrates how to:
// 1. Define tool parameters with JSON schema
// 2. Call other Liminal tools from within your tool
// 3. Process and analyze the data
// 4. Return useful insights
//
// Use this as a template for your own hackathon tools!

func createSpendingAnalyzerTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("analyze_spending").
		Description("Analyze the user's spending patterns over a specified time period. Returns insights about spending velocity, categories, and trends.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"days": tools.IntegerProperty("Number of days to analyze (default: 30)"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			// Parse input parameters
			var params struct {
				Days int `json:"days"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			// Default to 30 days if not specified
			if params.Days == 0 {
				params.Days = 30
			}

			// STEP 1: Fetch transaction history
			// We'll call the Liminal get_transactions tool through the executor
			txRequest := map[string]interface{}{
				"limit": 100, // Get up to 100 transactions
			}
			txRequestJSON, _ := json.Marshal(txRequest)

			txResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
				UserID:    toolParams.UserID,
				Tool:      "get_transactions",
				Input:     txRequestJSON,
				RequestID: toolParams.RequestID,
			})
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("failed to fetch transactions: %v", err),
				}, nil
			}

			if !txResponse.Success {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("transaction fetch failed: %s", txResponse.Error),
				}, nil
			}

			// STEP 2: Parse transaction data
			// In a real implementation, you'd parse the actual response structure
			// For now, we'll create a structured analysis

			var transactions []map[string]interface{}
			var txData map[string]interface{}
			if err := json.Unmarshal(txResponse.Data, &txData); err == nil {
				if txArray, ok := txData["transactions"].([]interface{}); ok {
					for _, tx := range txArray {
						if txMap, ok := tx.(map[string]interface{}); ok {
							transactions = append(transactions, txMap)
						}
					}
				}
			}

			// STEP 3: Analyze the data
			analysis := analyzeTransactions(transactions, params.Days)

			// STEP 4: Return insights
			result := map[string]interface{}{
				"period_days":        params.Days,
				"total_transactions": len(transactions),
				"analysis":           analysis,
				"generated_at":       time.Now().Format(time.RFC3339),
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

// analyzeTransactions processes transaction data and returns insights
func analyzeTransactions(transactions []map[string]interface{}, days int) map[string]interface{} {
	if len(transactions) == 0 {
		return map[string]interface{}{
			"summary": "No transactions found in the specified period",
		}
	}

	// Calculate basic metrics
	var totalSpent, totalReceived float64
	var spendCount, receiveCount int

	// This is a simplified example - you'd do real analysis here:
	// - Group by category/merchant
	// - Calculate daily/weekly averages
	// - Identify spending spikes
	// - Compare to previous periods
	// - Detect recurring payments

	for _, tx := range transactions {
		// Example analysis logic
		txType, _ := tx["type"].(string)
		amount, _ := tx["amount"].(float64)

		switch txType {
		case "send":
			totalSpent += amount
			spendCount++
		case "receive":
			totalReceived += amount
			receiveCount++
		}
	}

	avgDailySpend := totalSpent / float64(days)

	return map[string]interface{}{
		"total_spent":     fmt.Sprintf("%.2f", totalSpent),
		"total_received":  fmt.Sprintf("%.2f", totalReceived),
		"spend_count":     spendCount,
		"receive_count":   receiveCount,
		"avg_daily_spend": fmt.Sprintf("%.2f", avgDailySpend),
		"velocity":        calculateVelocity(spendCount, days),
		"insights": []string{
			fmt.Sprintf("You made %d spending transactions over %d days", spendCount, days),
			fmt.Sprintf("Average daily spend: $%.2f", avgDailySpend),
			"Consider setting up savings goals to build financial cushion",
		},
	}
}

// calculateVelocity determines spending frequency
func calculateVelocity(transactionCount, days int) string {
	txPerWeek := float64(transactionCount) / float64(days) * 7

	switch {
	case txPerWeek < 2:
		return "low"
	case txPerWeek < 7:
		return "moderate"
	default:
		return "high"
	}
}

// ============================================================================
// CUSTOM TOOL: WEEKLY SPENDING GOAL
// ============================================================================
// Sets and tracks weekly spending goals with progress monitoring

// In-memory storage for weekly goals (in production, use a database)
var weeklyGoals = make(map[string]map[string]interface{})

func createSpendWeeklyGoalTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("spend_weekly_goal").
		Description("Set or update a weekly spending goal. Extracts amount and currency from user input and tracks weekly spending progress.").
		RequiresConfirmation(). // Require user confirmation like WRITE OPERATIONS
		Schema(tools.ObjectSchema(map[string]interface{}{
			"amount":   tools.NumberProperty("The weekly spending limit amount"),
			"currency": tools.StringProperty("The currency code (e.g., USD, LIL, USDC)"),
			"action":   tools.StringProperty("Action: 'set' to create/update goal, 'get' to check current progress (default: set)"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				Amount   float64 `json:"amount"`
				Currency string  `json:"currency"`
				Action   string  `json:"action"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			// Default action is 'set'
			if params.Action == "" {
				params.Action = "set"
			}

			// Default currency
			if params.Currency == "" {
				params.Currency = "USD"
			}

			userID := toolParams.UserID

			// Handle GET action - check current progress
			if params.Action == "get" {
				return getWeeklySpendingProgress(ctx, liminalExecutor, toolParams)
			}

			// Handle SET action - create or update goal
			if params.Amount <= 0 {
				return &core.ToolResult{
					Success: false,
					Error:   "amount must be greater than 0",
				}, nil
			}

			// Get current week start (Monday)
			now := time.Now()
			weekStart := getWeekStart(now)
			weekEnd := weekStart.AddDate(0, 0, 7)

			// Store the goal
			weeklyGoals[userID] = map[string]interface{}{
				"amount":      params.Amount,
				"currency":    params.Currency,
				"week_start":  weekStart.Format("2006-01-02"),
				"week_end":    weekEnd.Format("2006-01-02"),
				"set_at":      now.Format(time.RFC3339),
			}

			// Get current spending for this week
			progress, err := calculateWeeklyProgress(ctx, liminalExecutor, toolParams, params.Amount, params.Currency)
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("failed to calculate progress: %v", err),
				}, nil
			}

			result := map[string]interface{}{
				"status":       "goal_set",
				"goal_set":     true,
				"goal_amount":  params.Amount,
				"currency":     params.Currency,
				"week_start":   weekStart.Format("Monday, Jan 2"),
				"week_end":     weekEnd.Format("Monday, Jan 2"),
				"spent_so_far": progress["spent"],
				"remaining":    progress["remaining"],
				"percentage":   progress["percentage"],
				"on_track":     progress["on_track"],
				"days_left":    progress["days_left"],
				"message":      fmt.Sprintf("Weekly spending goal set to %.2f %s", params.Amount, params.Currency),
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

// Helper function to get weekly spending progress
func getWeeklySpendingProgress(ctx context.Context, liminalExecutor core.ToolExecutor, toolParams *core.ToolParams) (*core.ToolResult, error) {
	userID := toolParams.UserID
	goal, exists := weeklyGoals[userID]

	if !exists {
		return &core.ToolResult{
			Success: true,
			Data: map[string]interface{}{
				"goal_set": false,
			},
		}, nil
	}

	amount := goal["amount"].(float64)
	currency := goal["currency"].(string)

	progress, err := calculateWeeklyProgress(ctx, liminalExecutor, toolParams, amount, currency)
	if err != nil {
		return &core.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to calculate progress: %v", err),
		}, nil
	}

	result := map[string]interface{}{
		"goal_set":     true,
		"goal_amount":  amount,
		"currency":     currency,
		"week_start":   goal["week_start"],
		"week_end":     goal["week_end"],
		"spent_so_far": progress["spent"],
		"remaining":    progress["remaining"],
		"percentage":   progress["percentage"],
		"on_track":     progress["on_track"],
		"days_left":    progress["days_left"],
	}

	return &core.ToolResult{
		Success: true,
		Data:    result,
	}, nil
}

// Calculate weekly spending progress
func calculateWeeklyProgress(ctx context.Context, liminalExecutor core.ToolExecutor, toolParams *core.ToolParams, goalAmount float64, currency string) (map[string]interface{}, error) {
	// Get transactions from this week
	txRequest := map[string]interface{}{
		"limit": 100,
	}
	txRequestJSON, _ := json.Marshal(txRequest)

	txResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
		UserID:    toolParams.UserID,
		Tool:      "get_transactions",
		Input:     txRequestJSON,
		RequestID: toolParams.RequestID,
	})
	if err != nil {
		return nil, err
	}

	if !txResponse.Success {
		return nil, fmt.Errorf("transaction fetch failed: %s", txResponse.Error)
	}
	
	log.Println("Transaction data:")
	log.Println(string(txResponse.Data))

	// Parse transactions
	var transactions []map[string]interface{}
	var txData map[string]interface{}
	if err := json.Unmarshal(txResponse.Data, &txData); err == nil {
		if txArray, ok := txData["transactions"].([]interface{}); ok {
			for _, tx := range txArray {
				if txMap, ok := tx.(map[string]interface{}); ok {
					transactions = append(transactions, txMap)
				}
			}
		}
	}
	// Calculate spending for this week using transaction dates
	weekStart := getWeekStart(time.Now())
	weekEnd := weekStart.AddDate(0, 0, 7)
	var weeklySpending float64

	log.Printf("Week range: %s to %s", weekStart.Format(time.RFC3339), weekEnd.Format(time.RFC3339))

	for _, tx := range transactions {
		// Parse amount (it's a string in the response)
		amountStr, _ := tx["amount"].(string)
		amount := 0.0
		fmt.Sscanf(amountStr, "%f", &amount)
		
		txCurrency, _ := tx["currency"].(string)
		direction, _ := tx["direction"].(string)
		
		// Parse createdAt timestamp (RFC3339 format)
		var txTime time.Time
		if createdAt, ok := tx["createdAt"].(string); ok {
			txTime, _ = time.Parse(time.RFC3339, createdAt)
		}
		
		log.Printf("Transaction: amount=%s (%.2f), currency=%s, direction=%s, date=%s, inWeek=%t", 
			amountStr, amount, txCurrency, direction, txTime.Format("2006-01-02"), 
			!txTime.IsZero() && (txTime.Equal(weekStart) || txTime.After(weekStart)) && txTime.Before(weekEnd))
		
		// Only count spending (debit/negative amounts) from this week in matching currency
		if !txTime.IsZero() && (txTime.Equal(weekStart) || txTime.After(weekStart)) && txTime.Before(weekEnd) {
			if txCurrency == currency || currency == "" {
				// Count debit transactions (money going out) or negative amounts
				if direction == "debit" || amount < 0 {
					weeklySpending += -amount // Make positive for display
				}
			}
		}
	}

	remaining := goalAmount - weeklySpending
	percentage := (weeklySpending / goalAmount) * 100
	if percentage > 100 {
		percentage = 100
	}

	now := time.Now()
	daysLeft := int(weekEnd.Sub(now).Hours() / 24)
	if daysLeft < 0 {
		daysLeft = 0
	}

	// Determine if on track
	dayOfWeek := int(now.Weekday())
	if dayOfWeek == 0 {
		dayOfWeek = 7 // Sunday is 7
	}
	expectedSpending := (goalAmount / 7) * float64(dayOfWeek)
	onTrack := weeklySpending <= expectedSpending

	log.Printf("Weekly goal progress: spent=%.2f remaining=%.2f percentage=%.2f on_track=%t days_left=%d currency=%s goal=%.2f", weeklySpending, remaining, percentage, onTrack, daysLeft, currency, goalAmount)

	return map[string]interface{}{
		"spent":      weeklySpending,
		"remaining":  remaining,
		"percentage": percentage,
		"on_track":   onTrack,
		"days_left":  daysLeft,
	}, nil
}

// Get the start of the current week (Monday)
func getWeekStart(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday is 7
	}
	daysToMonday := weekday - 1
	monday := t.AddDate(0, 0, -daysToMonday)
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
}

// ============================================================================
// CUSTOM TOOL: GET WEEKLY GOAL PROGRESS (READ-ONLY)
// ============================================================================
// Read-only version of weekly goal check that doesn't require confirmation

func createGetWeeklyGoalProgressTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("get_weekly_spending_progress").
		Description("Get current weekly spending goal progress without requiring confirmation. Shows how much spent, remaining budget, and on-track status.").
		Schema(tools.ObjectSchema(map[string]interface{}{})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			return getWeeklySpendingProgress(ctx, liminalExecutor, toolParams)
		}).
		Build()
}

// ============================================================================
// CUSTOM TOOL: CHECK WEEKLY SPEND (CONTEXT TOOL FOR AGENT)
// ============================================================================
// Provides agent with quick context about weekly spending status

func createCheckWeeklySpendTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("check_weeklyspend").
		Description("Check the current weekly spending status. Returns spent amount, remaining budget, percentage used, on-track status, and days left in the week. Use this to get context before answering user questions about their spending.").
		Schema(tools.ObjectSchema(map[string]interface{}{})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			userID := toolParams.UserID
			goal, exists := weeklyGoals[userID]

			if !exists {
				return &core.ToolResult{
					Success: true,
					Data: map[string]interface{}{
						"goal_set": false,
						"message":  "No weekly spending goal has been set yet",
					},
				}, nil
			}

			amount := goal["amount"].(float64)
			currency := goal["currency"].(string)

			progress, err := calculateWeeklyProgress(ctx, liminalExecutor, toolParams, amount, currency)
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("failed to calculate progress: %v", err),
				}, nil
			}

			result := map[string]interface{}{
				"goal_set":    true,
				"goal_amount": amount,
				"currency":    currency,
				"spent":       progress["spent"],
				"remaining":   progress["remaining"],
				"percentage":  progress["percentage"],
				"on_track":    progress["on_track"],
				"days_left":   progress["days_left"],
				"week_start":  goal["week_start"],
				"week_end":    goal["week_end"],
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

// ============================================================================
// CUSTOM TOOL: CALENDAR REMINDER
// ============================================================================
// Creates calendar reminders for periodic investment goals using Google Calendar API

// In-memory storage for calendar reminders (in production, use a database)
var calendarReminders = make(map[string]map[string]interface{})

func createCalendarReminderTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("create_calendar_reminder").
		Description("Create calendar reminders for periodic investments (weekly, bi-weekly, or monthly). This requires user confirmation before creating events.").
		RequiresConfirmation(). // Require user confirmation like WRITE OPERATIONS
		Schema(tools.ObjectSchema(map[string]interface{}{
			"frequency":  tools.StringProperty("Investment frequency: 'weekly', 'bi-weekly', or 'monthly'"),
			"amount":     tools.NumberProperty("Amount to invest per period"),
			"currency":   tools.StringProperty("Currency code (e.g., USDC, EURC)"),
			"start_date": tools.StringProperty("Start date for reminders (YYYY-MM-DD format, optional - defaults to next week)"),
			"duration":   tools.IntegerProperty("Number of reminders to create (default: 12 for weekly, 6 for bi-weekly, 3 for monthly)"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				Frequency string  `json:"frequency"`
				Amount    float64 `json:"amount"`
				Currency  string  `json:"currency"`
				StartDate string  `json:"start_date"`
				Duration  int     `json:"duration"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			// Validate frequency
			params.Frequency = strings.ToLower(params.Frequency)
			if params.Frequency != "weekly" && params.Frequency != "bi-weekly" && params.Frequency != "monthly" {
				return &core.ToolResult{
					Success: false,
					Error:   "frequency must be 'weekly', 'bi-weekly', or 'monthly'",
				}, nil
			}

			// Validate amount
			if params.Amount <= 0 {
				return &core.ToolResult{
					Success: false,
					Error:   "amount must be greater than 0",
				}, nil
			}

			// Default currency
			if params.Currency == "" {
				params.Currency = "USDC"
			}

			// Set default duration based on frequency
			if params.Duration == 0 {
				switch params.Frequency {
				case "weekly":
					params.Duration = 12 // 3 months of weekly reminders
				case "bi-weekly":
					params.Duration = 6 // 3 months of bi-weekly reminders
				case "monthly":
					params.Duration = 3 // 3 months of monthly reminders
				}
			}

			// Parse start date or default to next week
			var startDate time.Time
			if params.StartDate != "" {
				var err error
				startDate, err = time.Parse("2006-01-02", params.StartDate)
				if err != nil {
					return &core.ToolResult{
						Success: false,
						Error:   "start_date must be in YYYY-MM-DD format",
					}, nil
				}
			} else {
				// Default to next Monday
				startDate = getNextMonday()
			}

			// Create calendar events
			events, googleSyncSuccess, err := createCalendarEvents(params.Frequency, params.Amount, params.Currency, startDate, params.Duration)
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("failed to create calendar events: %v", err),
				}, nil
			}

			// Store reminder configuration
			userID := toolParams.UserID
			calendarReminders[userID] = map[string]interface{}{
				"frequency":        params.Frequency,
				"amount":           params.Amount,
				"currency":         params.Currency,
				"start_date":       startDate.Format("2006-01-02"),
				"duration":         params.Duration,
				"events":           events,
				"created_at":       time.Now().Format(time.RFC3339),
				"google_synced":    googleSyncSuccess,
			}

			// Build result message
			message := fmt.Sprintf("✅ Created %d calendar reminders for %s investing", params.Duration, params.Frequency)
			if googleSyncSuccess {
				message += "\n📅 Events synced to Google Calendar - you'll receive email notifications!"
			} else {
				message += "\n💾 Events stored locally (Google Calendar sync not configured)"
			}

			result := map[string]interface{}{
				"success":         true,
				"message":         message,
				"frequency":       params.Frequency,
				"amount":          params.Amount,
				"currency":        params.Currency,
				"start_date":      startDate.Format("January 2, 2006"),
				"next_reminder":   events[0]["date"],
				"total_events":    len(events),
				"google_synced":   googleSyncSuccess,
				"events":          events,
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

// Helper: Get next Monday
func getNextMonday() time.Time {
	now := time.Now()
	daysUntilMonday := (8 - int(now.Weekday())) % 7
	if daysUntilMonday == 0 {
		daysUntilMonday = 7
	}
	return now.AddDate(0, 0, daysUntilMonday)
}

// Helper: Create calendar events
func createCalendarEvents(frequency string, amount float64, currency string, startDate time.Time, duration int) ([]map[string]interface{}, bool, error) {
	events := make([]map[string]interface{}, 0)
	currentDate := startDate

	// Calculate interval based on frequency
	var intervalDays int
	var frequencyLabel string
	switch frequency {
	case "weekly":
		intervalDays = 7
		frequencyLabel = "Weekly"
	case "bi-weekly":
		intervalDays = 14
		frequencyLabel = "Bi-Weekly"
	case "monthly":
		intervalDays = 30
		frequencyLabel = "Monthly"
	}

	// Create events for each period
	for i := 0; i < duration; i++ {
		eventDate := currentDate.AddDate(0, 0, i*intervalDays)
		event := map[string]interface{}{
			"date":        eventDate.Format("2006-01-02"),
			"time":        "09:00", // 9 AM reminder
			"title":       fmt.Sprintf("%s Investment Reminder", frequencyLabel),
			"description": fmt.Sprintf("Time to invest %.2f %s into your savings vault. Stay on track with your financial goals!", amount, currency),
			"amount":      amount,
			"currency":    currency,
		}
		events = append(events, event)
	}

	// Try to create actual Google Calendar events (if credentials available)
	// This is optional - the tool works even without Google Calendar API credentials
	log.Printf("📅 Attempting to sync %d events to Google Calendar...", len(events))
	if err := createGoogleCalendarEvents(events); err != nil {
		log.Printf("⚠️  Could not create Google Calendar events: %v", err)
		log.Println("💡 Events are stored locally. To enable Google Calendar sync, add GOOGLE_CALENDAR_CREDENTIALS to .env")
		return events, false, nil // Return false for google sync status
	} else {
		log.Printf("✅ Successfully created %d events in Google Calendar!", len(events))
		return events, true, nil // Return true for google sync status
	}
}

// Helper: Create Google Calendar events (optional - requires credentials)
func createGoogleCalendarEvents(events []map[string]interface{}) error {
	// Check if Google Calendar credentials are available
	credsPath := os.Getenv("GOOGLE_CALENDAR_CREDENTIALS")
	if credsPath == "" {
		return fmt.Errorf("GOOGLE_CALENDAR_CREDENTIALS not set - skipping calendar sync")
	}

	log.Printf("🔑 Found credentials at: %s", credsPath)

	// Create calendar service
	ctx := context.Background()
	srv, err := calendar.NewService(ctx, option.WithCredentialsFile(credsPath))
	if err != nil {
		return fmt.Errorf("unable to create Calendar service: %v", err)
	}
	log.Println("✅ Successfully authenticated with Google Calendar API")

	// Create events in Google Calendar
	log.Printf("📝 Creating %d events...", len(events))
	for i, event := range events {
		dateStr := event["date"].(string)
		timeStr := event["time"].(string)
		startDateTime := fmt.Sprintf("%sT%s:00", dateStr, timeStr)

		calendarEvent := &calendar.Event{
			Summary:     event["title"].(string),
			Description: event["description"].(string),
			Start: &calendar.EventDateTime{
				DateTime: startDateTime,
				TimeZone: "America/New_York",
			},
			End: &calendar.EventDateTime{
				DateTime: startDateTime, // 0-duration event (reminder)
				TimeZone: "America/New_York",
			},
			Reminders: &calendar.EventReminders{
				UseDefault: false,
				Overrides: []*calendar.EventReminder{
					{Method: "popup", Minutes: 0},
					{Method: "email", Minutes: 60}, // 1 hour before
				},
			},
		}

		createdEvent, err := srv.Events.Insert("primary", calendarEvent).Do()
		if err != nil {
			return fmt.Errorf("unable to create event %d: %v", i+1, err)
		}
		log.Printf("  ✓ Event %d/%d: %s on %s (ID: %s)", i+1, len(events), event["title"], event["date"], createdEvent.Id)
	}

	return nil
}

// ============================================================================
// CUSTOM TOOL: CATEGORIZE TRANSACTIONS
// ============================================================================
// Analyzes transaction notes and categorizes spending patterns using Claude structured output

func createCategorizeTransactionTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("categorize_transactions").
		Description("Analyze transaction notes and categorize spending into: food, travel, subscription, entertainment, electronics, miscellaneous using AI-powered categorization.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"limit": tools.IntegerProperty("Number of transactions to analyze (default: 50)"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				Limit int `json:"limit"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			if params.Limit == 0 {
				params.Limit = 50
			}

			// Fetch transactions
			txRequest := map[string]interface{}{"limit": params.Limit}
			txRequestJSON, _ := json.Marshal(txRequest)

			txResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
				UserID:    toolParams.UserID,
				Tool:      "get_transactions",
				Input:     txRequestJSON,
				RequestID: toolParams.RequestID,
			})
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("failed to fetch transactions: %v", err),
				}, nil
			}

			if !txResponse.Success {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("transaction fetch failed: %s", txResponse.Error),
				}, nil
			}

			// Parse transactions
			var transactions []map[string]interface{}
			var txData map[string]interface{}
			if err := json.Unmarshal(txResponse.Data, &txData); err == nil {
				if txArray, ok := txData["transactions"].([]interface{}); ok {
					for _, tx := range txArray {
						if txMap, ok := tx.(map[string]interface{}); ok {
							transactions = append(transactions, txMap)
						}
					}
				}
			}

			// Extract spending transaction notes
			var spendingNotes []string
			for _, tx := range transactions {
				direction, _ := tx["direction"].(string)
				amountStr, _ := tx["amount"].(string)
				amount := 0.0
				fmt.Sscanf(amountStr, "%f", &amount)

				// Only categorize spending (debit or negative)
				if direction == "debit" || amount < 0 {
					if note, ok := tx["note"].(string); ok && note != "" {
						spendingNotes = append(spendingNotes, note)
					}
				}
			}

			if len(spendingNotes) == 0 {
				return &core.ToolResult{
					Success: true,
					Data: map[string]interface{}{
						"categories": map[string]int{
							"food":          0,
							"travel":        0,
							"subscription":  0,
							"entertainment": 0,
							"electronics":   0,
							"miscellaneous": 0,
						},
						"total_analyzed": 0,
						"breakdown":      []string{},
					},
				}, nil
			}

			// Use Claude structured output to categorize
			categorized, err := categorizeWithStructuredOutput(spendingNotes)
			if err != nil {
				log.Printf("AI categorization failed, using fallback: %v", err)
				// Fallback to keyword matching
				categorized = fallbackCategorization(spendingNotes)
			}

			return &core.ToolResult{
				Success: true,
				Data:    categorized,
			}, nil
		}).
		Build()
}

// fallbackCategorization uses keyword matching when AI categorization fails
func fallbackCategorization(notes []string) map[string]interface{} {
	categories := map[string]int{
		"food":          0,
		"travel":        0,
		"subscription":  0,
		"entertainment": 0,
		"electronics":   0,
		"miscellaneous": 0,
	}

	var breakdown []string
	for _, note := range notes {
		category := categorizeNote(note)
		categories[category]++
		breakdown = append(breakdown, fmt.Sprintf("%s: %s", note, category))
	}

	return map[string]interface{}{
		"categories":     categories,
		"total_analyzed": len(notes),
		"breakdown":      breakdown,
	}
}

// categorizeNote uses simple keyword matching to categorize transaction notes
func categorizeNote(note string) string {
	note = strings.ToLower(note)

	// Food keywords
	if strings.Contains(note, "food") || strings.Contains(note, "restaurant") ||
		strings.Contains(note, "cafe") || strings.Contains(note, "coffee") ||
		strings.Contains(note, "lunch") || strings.Contains(note, "dinner") ||
		strings.Contains(note, "grocery") || strings.Contains(note, "meal") ||
		strings.Contains(note, "hulu") {
		return "food"
	}

	// Travel keywords
	if strings.Contains(note, "uber") || strings.Contains(note, "lyft") ||
		strings.Contains(note, "flight") || strings.Contains(note, "hotel") ||
		strings.Contains(note, "gas") || strings.Contains(note, "parking") ||
		strings.Contains(note, "taxi") || strings.Contains(note, "bus") ||
		strings.Contains(note, "train") || strings.Contains(note, "ticket") {
		return "travel"
	}

	// Subscription keywords
	if strings.Contains(note, "subscription") || strings.Contains(note, "netflix") ||
		strings.Contains(note, "spotify") || strings.Contains(note, "monthly") ||
		strings.Contains(note, "membership") || strings.Contains(note, "premium") {
		return "subscription"
	}

	// Entertainment keywords
	if strings.Contains(note, "movie") || strings.Contains(note, "concert") ||
		strings.Contains(note, "game") || strings.Contains(note, "entertainment") ||
		strings.Contains(note, "video") {
		return "entertainment"
	}

	// Electronics keywords
	if strings.Contains(note, "phone") || strings.Contains(note, "laptop") ||
		strings.Contains(note, "computer") || strings.Contains(note, "electronics") ||
		strings.Contains(note, "gadget") || strings.Contains(note, "tech") {
		return "electronics"
	}

	return "miscellaneous"
}

// categorizeWithStructuredOutput uses Claude's structured output to categorize transactions
func categorizeWithStructuredOutput(notes []string) (map[string]interface{}, error) {
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey == "" {
		return nil, fmt.Errorf("missing ANTHROPIC_API_KEY")
	}

	// Define structured output schema
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"categorized_transactions": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"note": map[string]interface{}{
							"type": "string",
						},
						"category": map[string]interface{}{
							"type": "string",
							"enum": []string{"food", "travel", "subscription", "entertainment", "electronics", "miscellaneous"},
						},
					},
					"required": []string{"note", "category"},
				},
			},
		},
		"required": []string{"categorized_transactions"},
	}

	// Build prompt with transaction notes
	prompt := "Categorize the following transaction notes into one of these categories: food, travel, subscription, entertainment, electronics, or miscellaneous.\n\nTransaction notes:\n"
	for i, note := range notes {
		prompt += fmt.Sprintf("%d. %s\n", i+1, note)
	}

	// Call Claude API with structured output
	requestBody := map[string]interface{}{
		"model":      "claude-sonnet-4-20250514",
		"max_tokens": 2048,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"response_format": map[string]interface{}{
			"type":        "json_schema",
			"json_schema": schema,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", anthropicKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	var apiResponse struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	if len(apiResponse.Content) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	// Parse structured output
	var structured struct {
		CategorizedTransactions []struct {
			Note     string `json:"note"`
			Category string `json:"category"`
		} `json:"categorized_transactions"`
	}

	if err := json.Unmarshal([]byte(apiResponse.Content[0].Text), &structured); err != nil {
		return nil, err
	}

	// Build result
	categories := map[string]int{
		"food":          0,
		"travel":        0,
		"subscription":  0,
		"entertainment": 0,
		"electronics":   0,
		"miscellaneous": 0,
	}

	var breakdown []string
	for _, item := range structured.CategorizedTransactions {
		categories[item.Category]++
		breakdown = append(breakdown, fmt.Sprintf("%s: %s", item.Note, item.Category))
	}

	return map[string]interface{}{
		"categories":     categories,
		"total_analyzed": len(structured.CategorizedTransactions),
		"breakdown":      breakdown,
	}, nil
}

// generateBalanceTrendFromTransactions creates a line chart showing balance over time
func generateBalanceTrendFromTransactions(transactions []map[string]interface{}, currentBalance float64, days int) map[string]interface{} {
	log.Printf("📊 Generating balance trend from %d transactions, current balance: %.2f", len(transactions), currentBalance)

	if len(transactions) == 0 {
		return map[string]interface{}{
			"labels": []string{"Today"},
			"values": []float64{currentBalance},
			"title":  "Account Balance Trend",
		}
	}

	// Sort transactions by date (oldest first)
	sortedTxs := make([]map[string]interface{}, len(transactions))
	copy(sortedTxs, transactions)
	
	sort.Slice(sortedTxs, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, sortedTxs[i]["createdAt"].(string))
		timeJ, _ := time.Parse(time.RFC3339, sortedTxs[j]["createdAt"].(string))
		return timeI.Before(timeJ)
	})

	// Calculate starting balance by working backwards from current balance
	runningBalance := currentBalance
	for i := len(sortedTxs) - 1; i >= 0; i-- {
		tx := sortedTxs[i]
		direction, _ := tx["direction"].(string)
		
		var amount float64
		if amountStr, ok := tx["amount"].(string); ok {
			fmt.Sscanf(amountStr, "%f", &amount)
		} else if amountFloat, ok := tx["amount"].(float64); ok {
			amount = amountFloat
		}

		// Reverse the transaction to get starting balance
		if direction == "credit" {
			runningBalance -= amount // Was a credit, so subtract to go back
		} else if direction == "debit" {
			runningBalance += amount // Was a debit, so add to go back
		}
	}

	startingBalance := runningBalance
	log.Printf("📊 Calculated starting balance: %.2f", startingBalance)

	// Now go forward through transactions building the balance timeline
	labels := []string{}
	values := []float64{}
	
	// Add starting point
	if len(sortedTxs) > 0 {
		firstTxTime, _ := time.Parse(time.RFC3339, sortedTxs[0]["createdAt"].(string))
		labels = append(labels, firstTxTime.Format("Jan 2"))
		values = append(values, startingBalance)
	}

	runningBalance = startingBalance
	for _, tx := range sortedTxs {
		direction, _ := tx["direction"].(string)
		
		var amount float64
		if amountStr, ok := tx["amount"].(string); ok {
			fmt.Sscanf(amountStr, "%f", &amount)
		} else if amountFloat, ok := tx["amount"].(float64); ok {
			amount = amountFloat
		}

		// Apply transaction
		if direction == "credit" {
			runningBalance += amount
		} else if direction == "debit" {
			runningBalance -= amount
		}

		txTime, _ := time.Parse(time.RFC3339, tx["createdAt"].(string))
		labels = append(labels, txTime.Format("Jan 2"))
		values = append(values, runningBalance)
		
		log.Printf("📊 %s: %s $%.2f -> Balance: $%.2f", txTime.Format("Jan 2"), direction, amount, runningBalance)
	}

	log.Printf("📊 Final chart: %d data points", len(labels))

	return map[string]interface{}{
		"labels": labels,
		"values": values,
		"title":  "Account Balance Trend",
	}
}

// generateSVGChart creates an SVG image from chart data
func generateSVGChart(chartData map[string]interface{}) string {
	labels, _ := chartData["labels"].([]string)
	valuesInterface, _ := chartData["values"].([]float64)
	title, _ := chartData["title"].(string)

	if len(labels) == 0 || len(valuesInterface) == 0 {
		return `<svg width="600" height="400" xmlns="http://www.w3.org/2000/svg"><text x="300" y="200" text-anchor="middle" fill="#666">No data available</text></svg>`
	}

	// Chart dimensions
	width := 800
	height := 500
	padding := 80
	chartWidth := width - padding*2
	chartHeight := height - padding*2

	// Find min/max values
	minValue := valuesInterface[0]
	maxValue := valuesInterface[0]
	for _, v := range valuesInterface {
		if v < minValue {
			minValue = v
		}
		if v > maxValue {
			maxValue = v
		}
	}
	valueRange := maxValue - minValue
	if valueRange == 0 {
		valueRange = 1
	}

	// Build SVG
	var svg strings.Builder
	svg.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, width, height))
	
	// Background
	svg.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="#ffffff"/>`, width, height))
	
	// Title
	svg.WriteString(fmt.Sprintf(`<text x="%d" y="30" text-anchor="middle" font-size="20" font-weight="bold" fill="#333">%s</text>`, width/2, title))
	
	// Grid lines and Y-axis labels
	for i := 0; i <= 4; i++ {
		y := float64(padding) + float64(chartHeight*i)/4
		gridValue := maxValue - (float64(i)/4)*valueRange
		svg.WriteString(fmt.Sprintf(`<line x1="%d" y1="%.1f" x2="%d" y2="%.1f" stroke="#e0e0e0" stroke-width="1"/>`, padding, y, width-padding, y))
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%.1f" text-anchor="end" font-size="12" fill="#666">$%.0f</text>`, padding-10, y+4, gridValue))
	}
	
	// Build points and line path
	var points []string
	for i, value := range valuesInterface {
		x := float64(padding) + (float64(i)/float64(len(valuesInterface)-1))*float64(chartWidth)
		y := float64(padding) + float64(chartHeight) - ((value-minValue)/valueRange)*float64(chartHeight)
		points = append(points, fmt.Sprintf("%.1f,%.1f", x, y))
	}
	
	// Draw line
	svg.WriteString(fmt.Sprintf(`<polyline points="%s" fill="none" stroke="#4ECDC4" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>`, strings.Join(points, " ")))
	
	// Draw points and labels
	labelStep := 1
	if len(labels) > 15 {
		labelStep = len(labels) / 10
	}
	
	for i, value := range valuesInterface {
		x := float64(padding) + (float64(i)/float64(len(valuesInterface)-1))*float64(chartWidth)
		y := float64(padding) + float64(chartHeight) - ((value-minValue)/valueRange)*float64(chartHeight)
		
		// Draw point
		svg.WriteString(fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="4" fill="#4ECDC4" stroke="white" stroke-width="2"/>`, x, y))
		
		// Draw label (only for selected points to avoid crowding)
		if i%labelStep == 0 || i == len(labels)-1 {
			svg.WriteString(fmt.Sprintf(`<text x="%.1f" y="%d" text-anchor="middle" font-size="10" fill="#666" transform="rotate(-45 %.1f %d)">%s</text>`, x, height-padding+20, x, height-padding+20, labels[i]))
		}
	}
	
	svg.WriteString(`</svg>`)
	return svg.String()
}

// Generate investment comparison chart: Lump Sum vs Dollar-Cost Averaging
func generateInvestmentComparisonChart(principal float64, apy float64, currency string) string {
	// Calculate 12-month projections for both strategies
	months := 12
	
	// Lump sum: invest all upfront, earn full APY
	lumpSumValues := make([]float64, months+1)
	lumpSumValues[0] = principal
	monthlyRate := apy / 100 / 12
	for i := 1; i <= months; i++ {
		lumpSumValues[i] = lumpSumValues[i-1] * (1 + monthlyRate)
	}
	
	// DCA: invest 1/12 each month, each deposit earns proportionally less time
	dcaValues := make([]float64, months+1)
	dcaValues[0] = 0
	monthlyDeposit := principal / 12
	for i := 1; i <= months; i++ {
		// Add new deposit and compound existing balance
		dcaValues[i] = (dcaValues[i-1] + monthlyDeposit) * (1 + monthlyRate)
	}
	
	// Chart dimensions
	width := 800
	height := 500
	padding := 80
	chartWidth := width - 2*padding
	chartHeight := height - 2*padding
	
	// Find max value for scaling
	maxValue := lumpSumValues[months]
	minValue := 0.0
	valueRange := maxValue - minValue
	
	var svg strings.Builder
	svg.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, width, height))
	svg.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="#ffffff"/>`, width, height))
	
	// Title
	svg.WriteString(fmt.Sprintf(`<text x="%d" y="30" text-anchor="middle" font-size="18" font-weight="bold" fill="#333">Investment Strategy Comparison (%.2f%% APY)</text>`, width/2, apy))
	svg.WriteString(fmt.Sprintf(`<text x="%d" y="50" text-anchor="middle" font-size="14" fill="#666">Starting Amount: %.2f %s over 12 months</text>`, width/2, principal, currency))
	
	// Grid lines and Y-axis
	for i := 0; i <= 4; i++ {
		y := float64(padding) + float64(chartHeight*i)/4
		gridValue := maxValue - (float64(i)/4)*valueRange
		svg.WriteString(fmt.Sprintf(`<line x1="%d" y1="%.1f" x2="%d" y2="%.1f" stroke="#e0e0e0" stroke-width="1"/>`, padding, y, width-padding, y))
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%.1f" text-anchor="end" font-size="12" fill="#666">$%.0f</text>`, padding-10, y+4, gridValue))
	}
	
	// X-axis labels (months)
	for i := 0; i <= months; i += 3 {
		x := float64(padding) + (float64(i)/float64(months))*float64(chartWidth)
		svg.WriteString(fmt.Sprintf(`<text x="%.1f" y="%d" text-anchor="middle" font-size="12" fill="#666">Month %d</text>`, x, height-padding+25, i))
	}
	
	// Plot Lump Sum line (blue)
	var lumpSumPoints []string
	for i := 0; i <= months; i++ {
		x := float64(padding) + (float64(i)/float64(months))*float64(chartWidth)
		y := float64(padding) + float64(chartHeight) - ((lumpSumValues[i]-minValue)/valueRange)*float64(chartHeight)
		lumpSumPoints = append(lumpSumPoints, fmt.Sprintf("%.1f,%.1f", x, y))
	}
	svg.WriteString(fmt.Sprintf(`<polyline points="%s" fill="none" stroke="#2196F3" stroke-width="3"/>`, strings.Join(lumpSumPoints, " ")))
	
	// Plot DCA line (green)
	var dcaPoints []string
	for i := 0; i <= months; i++ {
		x := float64(padding) + (float64(i)/float64(months))*float64(chartWidth)
		y := float64(padding) + float64(chartHeight) - ((dcaValues[i]-minValue)/valueRange)*float64(chartHeight)
		dcaPoints = append(dcaPoints, fmt.Sprintf("%.1f,%.1f", x, y))
	}
	svg.WriteString(fmt.Sprintf(`<polyline points="%s" fill="none" stroke="#4CAF50" stroke-width="3" stroke-dasharray="5,5"/>`, strings.Join(dcaPoints, " ")))
	
	// Legend
	legendX := padding + 20
	legendY := padding + 20
	svg.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#2196F3" stroke-width="3"/>`, legendX, legendY, legendX+40, legendY))
	svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="14" fill="#333">💰 Lump Sum: $%.2f</text>`, legendX+50, legendY+5, lumpSumValues[months]))
	
	svg.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#4CAF50" stroke-width="3" stroke-dasharray="5,5"/>`, legendX, legendY+25, legendX+40, legendY+25))
	svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="14" fill="#333">📅 DCA (Chunks): $%.2f</text>`, legendX+50, legendY+30, dcaValues[months]))
	
	// Difference annotation
	difference := lumpSumValues[months] - dcaValues[months]
	svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="12" fill="#FF5722">Difference: $%.2f (%.1f%%)</text>`, legendX, legendY+55, difference, (difference/dcaValues[months])*100))
	
	svg.WriteString(`</svg>`)
	return svg.String()
}

// FlaggedItem represents a spending item that's been flagged
type FlaggedItem struct {
	Category string
	Amount   float64
	Count    int
	Reason   string
}

// FlaggedSpending contains categorized flagged expenses
type FlaggedSpending struct {
	Unnecessary      []FlaggedItem
	Excessive        []FlaggedItem
	TotalUnnecessary float64
	TotalExcessive   float64
}

// Analyze transactions and flag unnecessary/excessive spending
func analyzeAndFlagSpending(transactions []interface{}, estimatedMonthlyIncome float64) FlaggedSpending {
	result := FlaggedSpending{
		Unnecessary: []FlaggedItem{},
		Excessive:   []FlaggedItem{},
	}
	
	// Category spending totals
	categorySpending := make(map[string]float64)
	categoryCounts := make(map[string]int)
	
	// Categorize each transaction
	for _, tx := range transactions {
		txMap, ok := tx.(map[string]interface{})
		if !ok {
			continue
		}
		
		direction, _ := txMap["direction"].(string)
		if direction != "debit" {
			continue // Only look at debits (spending)
		}
		
		note, _ := txMap["note"].(string)
		var amount float64
		if amountStr, ok := txMap["amount"].(string); ok {
			fmt.Sscanf(amountStr, "%f", &amount)
		}
		
		category := categorizeSingleNote(note)
		categorySpending[category] += amount
		categoryCounts[category]++
	}
	
	// Define thresholds (percentage of estimated income)
	unnecessaryCategories := map[string]bool{
		"entertainment": true,
		"subscription":  true,
	}
	
	essentialThresholds := map[string]float64{
		"food":        estimatedMonthlyIncome * 0.15, // 15% max
		"travel":      estimatedMonthlyIncome * 0.10, // 10% max
		"electronics": estimatedMonthlyIncome * 0.05, // 5% max
	}
	
	// Flag unnecessary spending
	for category, spending := range categorySpending {
		if unnecessaryCategories[category] && spending > 10 {
			result.Unnecessary = append(result.Unnecessary, FlaggedItem{
				Category: category,
				Amount:   spending,
				Count:    categoryCounts[category],
				Reason:   "Non-essential, consider cutting back",
			})
			result.TotalUnnecessary += spending
		}
		
		// Flag excessive essential spending
		if threshold, exists := essentialThresholds[category]; exists && spending > threshold {
			excessAmount := spending - threshold
			result.Excessive = append(result.Excessive, FlaggedItem{
				Category: category,
				Amount:   spending,
				Count:    categoryCounts[category],
				Reason:   fmt.Sprintf("%.0f%% over recommended budget", (excessAmount/threshold)*100),
			})
			result.TotalExcessive += excessAmount
		}
	}
	
	return result
}

// Generate flagged spending visualization (bubble chart style)
func generateFlaggedSpendingChart(flagged FlaggedSpending) string {
	width := 800
	height := 600
	
	var svg strings.Builder
	svg.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, width, height))
	svg.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="#ffffff"/>`, width, height))
	
	// Title
	svg.WriteString(`<text x="400" y="30" text-anchor="middle" font-size="20" font-weight="bold" fill="#333">Flagged Spending Analysis</text>`)
	svg.WriteString(`<text x="400" y="55" text-anchor="middle" font-size="14" fill="#666">Red = Unnecessary | Orange = Excessive</text>`)
	
	// Calculate bubble positions
	allItems := append([]FlaggedItem{}, flagged.Unnecessary...)
	allItems = append(allItems, flagged.Excessive...)
	
	if len(allItems) == 0 {
		svg.WriteString(`<text x="400" y="300" text-anchor="middle" font-size="18" fill="#4CAF50">✅ No major spending issues found!</text>`)
		svg.WriteString(`</svg>`)
		return svg.String()
	}
	
	// Find max amount for scaling
	maxAmount := 0.0
	for _, item := range allItems {
		if item.Amount > maxAmount {
			maxAmount = item.Amount
		}
	}
	
	// Draw bubbles
	y := 120
	rowHeight := 140
	
	for i, item := range flagged.Unnecessary {
		x := 150 + (i%3)*250
		if i > 0 && i%3 == 0 {
			y += rowHeight
		}
		currentY := y + (i/3)*rowHeight
		
		radius := 30 + (item.Amount/maxAmount)*40
		
		// Red bubble for unnecessary
		svg.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="%.1f" fill="#ff5252" opacity="0.7" stroke="#d32f2f" stroke-width="2"/>`, x, currentY, radius))
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" font-size="14" font-weight="bold" fill="#fff">%s</text>`, x, currentY-5, strings.Title(item.Category)))
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" font-size="16" font-weight="bold" fill="#fff">$%.0f</text>`, x, currentY+15, item.Amount))
	}
	
	// Continue with excessive items
	startIdx := len(flagged.Unnecessary)
	for i, item := range flagged.Excessive {
		idx := startIdx + i
		x := 150 + (idx%3)*250
		currentY := y + (idx/3)*rowHeight
		
		radius := 30 + (item.Amount/maxAmount)*40
		
		// Orange bubble for excessive
		svg.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="%.1f" fill="#ff9800" opacity="0.7" stroke="#f57c00" stroke-width="2"/>`, x, currentY, radius))
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" font-size="14" font-weight="bold" fill="#fff">%s</text>`, x, currentY-5, strings.Title(item.Category)))
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" font-size="16" font-weight="bold" fill="#fff">$%.0f</text>`, x, currentY+15, item.Amount))
	}
	
	// Summary at bottom
	summaryY := height - 80
	totalFlagged := flagged.TotalUnnecessary + flagged.TotalExcessive
	svg.WriteString(fmt.Sprintf(`<text x="400" y="%d" text-anchor="middle" font-size="18" font-weight="bold" fill="#333">Potential Monthly Savings: $%.2f</text>`, summaryY, totalFlagged))
	svg.WriteString(fmt.Sprintf(`<text x="400" y="%d" text-anchor="middle" font-size="14" fill="#666">Unnecessary: $%.2f | Excessive: $%.2f</text>`, summaryY+25, flagged.TotalUnnecessary, flagged.TotalExcessive))
	
	svg.WriteString(`</svg>`)
	return svg.String()
}

// Helper: Categorize a single transaction note
func categorizeSingleNote(note string) string {
	note = strings.ToLower(note)
	
	// Entertainment
	if strings.Contains(note, "game") || strings.Contains(note, "movie") || 
	   strings.Contains(note, "concert") || strings.Contains(note, "entertainment") {
		return "entertainment"
	}
	
	// Subscriptions
	if strings.Contains(note, "subscription") || strings.Contains(note, "netflix") ||
	   strings.Contains(note, "spotify") || strings.Contains(note, "hulu") ||
	   strings.Contains(note, "prime") {
		return "subscription"
	}
	
	// Food
	if strings.Contains(note, "food") || strings.Contains(note, "restaurant") ||
	   strings.Contains(note, "lunch") || strings.Contains(note, "dinner") ||
	   strings.Contains(note, "coffee") || strings.Contains(note, "groceries") {
		return "food"
	}
	
	// Travel
	if strings.Contains(note, "travel") || strings.Contains(note, "flight") ||
	   strings.Contains(note, "hotel") || strings.Contains(note, "uber") ||
	   strings.Contains(note, "taxi") || strings.Contains(note, "bus") ||
	   strings.Contains(note, "train") || strings.Contains(note, "gas") {
		return "travel"
	}
	
	// Electronics
	if strings.Contains(note, "phone") || strings.Contains(note, "laptop") ||
	   strings.Contains(note, "computer") || strings.Contains(note, "electronics") ||
	   strings.Contains(note, "gadget") || strings.Contains(note, "tech") {
		return "electronics"
	}
	
	return "miscellaneous"
}

// ============================================================================
// CUSTOM TOOL: GRAPH ORCHESTRATOR
// ============================================================================
// Routes requests through the LangGraph workflow system

func createGraphOrchestratorTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("route_request").
		Description("Analyze user's request and route to the appropriate specialized handler. Call this FIRST for any user request to determine the best way to help them. Returns routing decision and context.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"user_message": tools.StringProperty("The user's original message/request"),
		}, "user_message")).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				UserMessage string `json:"user_message"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			// Create graph and initialize state
			graph := CreateFinancialAgentGraph(liminalExecutor)
			graph.State = &GraphState{
				Messages:     []string{},
				UserID:       toolParams.UserID,
				Conversation: map[string]interface{}{
					"user_input": params.UserMessage,
				},
			}

			// Execute graph workflow
			if err := graph.Execute(ctx, toolParams.UserID); err != nil {
				log.Printf("Graph execution error: %v", err)
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("routing failed: %v", err),
				}, nil
			}

			// Extract routing decision
			route, _ := graph.State.Conversation["route"].(string)
			handlerType, _ := graph.State.Conversation["handler_type"].(string)

			var guidance string
			switch handlerType {
			case "image_payment":
				guidance = "User wants to split a payment or send money based on an image/receipt. You should ask for the receipt image, analyze it, calculate splits, and use send_money tool to process payments."
			case "withdraw":
				guidance = "User wants to withdraw money from savings. The system has analyzed their liquidity situation and provided educational content about withdrawal safety. Now help them complete the withdrawal using the withdraw_savings tool. Require: amount (as string) and currency ('USD' or 'EUR'). If this was an unsafe withdrawal, also offer to set up the weekly budget using spend_weekly_goal tool after completing the withdrawal."
			case "deposit":
				guidance = "User wants to deposit money into savings. The system has shown their available balances and potential earnings. Help them complete the deposit using the deposit_savings tool. Require: amount (as string) and currency ('USD' or 'EUR'). This tool requires user confirmation. Encourage them about the benefits of earning passive income through compound interest."
			case "financial_help":
				guidance = "User needs financial assistance with low funds. Use check_weeklyspend, analyze_spending, and categorize_transactions to help them improve their financial situation. Focus on budget management, reducing expenses, and building up savings."
			case "financial_save":
				// Get recommended savings details from graph state
				bestCurrency := "USDC"
				bestAPY := 4.0
				availableToSave := 0.0
				
				if result, ok := graph.State.ToolResult.(map[string]interface{}); ok {
					if curr, ok := result["best_currency"].(string); ok {
						bestCurrency = curr
					}
					if apy, ok := result["best_apy"].(float64); ok {
						bestAPY = apy
					}
					if available, ok := result["available_to_save"].(float64); ok {
						availableToSave = available
					}
				}
				
				guidance = fmt.Sprintf("User has sufficient funds for saving/investing. Based on vault rate analysis:\n\n"+
					"BEST OPTION: Save in %s vault earning %.2f%% APY\n"+
					"AVAILABLE: %.2f %s can be moved to savings\n\n"+
					"Provide personalized advice on:\n"+
					"1. The benefits of their recommended vault (%s at %.2f%% APY)\n"+
					"2. Investment strategy options:\n"+
					"   - Lump Sum: Deposit all at once (pros: immediate full interest, simpler; cons: timing risk)\n"+
					"   - Dollar-Cost Averaging: Deposit in chunks over time (pros: reduces timing risk, builds habit; cons: less immediate interest)\n"+
					"3. Suggest specific amounts based on their %.2f %s balance\n"+
					"4. Use deposit_savings tool if they want to proceed\n\n"+
					"Be encouraging and educational about building wealth through savings!",
					bestCurrency, bestAPY, availableToSave, bestCurrency, bestCurrency, bestAPY, availableToSave, bestCurrency)
			case "general":
				// Check if chart was requested
				if chartRequested, ok := graph.State.Conversation["chart_requested"].(bool); ok && chartRequested {
					guidance = "User requested a balance trend chart. You MUST call the generate_chart tool with these parameters:\n- chart_type: 'line'\n- data_type: 'balance_trend'\n- days: 30 (or ask the user for a timeframe)\n\nAfter calling generate_chart, you will receive an 'image_url' field with a base64-encoded SVG. Display it in your response using markdown:\n\n![Balance Trend Chart](image_url_here)\n\nReplace 'image_url_here' with the actual image_url from the tool result. Explain that this shows their account balance over time based on transaction history."
				} else {
					guidance = "Standard query. Use appropriate banking tools (get_balance, get_transactions, etc.) to help the user."
				}
			default:
				guidance = "Process as a general query."
			}

			result := map[string]interface{}{
				"route":        route,
				"handler_type": handlerType,
				"guidance":     guidance,
				"user_message": params.UserMessage,
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

// ============================================================================
// CUSTOM TOOL: CHART GENERATOR
// ============================================================================
// Generates charts and graphs from financial data using go-chart library
// Similar to matplotlib in Python

func createChartGeneratorTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("generate_chart").
		Description("Generate a line chart showing account balance trend over time. Calculates running balance from transaction history in chronological order.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"chart_type": tools.StringProperty("Type of chart: always 'line' for balance trend"),
			"data_type":  tools.StringProperty("What to visualize: always 'balance_trend'"),
			"days":       tools.IntegerProperty("Number of days of data to include (default: 30)"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				ChartType string `json:"chart_type"`
				DataType  string `json:"data_type"`
				Days      int    `json:"days"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			// Set defaults
			if params.Days == 0 {
				params.Days = 30
			}

			// Fetch transaction data
			txRequest := map[string]interface{}{"limit": 200}
			txRequestJSON, _ := json.Marshal(txRequest)

			txResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
				UserID:    toolParams.UserID,
				Tool:      "get_transactions",
				Input:     txRequestJSON,
				RequestID: toolParams.RequestID,
			})
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("failed to fetch transactions: %v", err),
				}, nil
			}

			if !txResponse.Success {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("transaction fetch failed: %s", txResponse.Error),
				}, nil
			}

			// Get current balance
			balanceResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
				UserID:    toolParams.UserID,
				Tool:      "get_balance",
				Input:     []byte("{}"),
				RequestID: toolParams.RequestID,
			})
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("failed to fetch balance: %v", err),
				}, nil
			}

			// Parse current balance
			var currentBalance float64
			if balanceResponse.Success {
				var balanceData map[string]interface{}
				if err := json.Unmarshal(balanceResponse.Data, &balanceData); err == nil {
					if balanceStr, ok := balanceData["balance"].(string); ok {
						fmt.Sscanf(balanceStr, "%f", &currentBalance)
					} else if balanceFloat, ok := balanceData["balance"].(float64); ok {
						currentBalance = balanceFloat
					}
				}
			}

			// Parse transactions
			var transactions []map[string]interface{}
			var txData map[string]interface{}
			if err := json.Unmarshal(txResponse.Data, &txData); err == nil {
				if txArray, ok := txData["transactions"].([]interface{}); ok {
					for _, tx := range txArray {
						if txMap, ok := tx.(map[string]interface{}); ok {
							transactions = append(transactions, txMap)
						}
					}
				}
			}

			// Generate balance trend chart
			chartData := generateBalanceTrendFromTransactions(transactions, currentBalance, params.Days)

			// Generate SVG image
			svgContent := generateSVGChart(chartData)
			
			// Save to charts directory
			chartsDir := filepath.Join(".", "charts")
			timestamp := time.Now().Format("20060102-150405")
			filename := fmt.Sprintf("balance-trend-%s.svg", timestamp)
			filePath := filepath.Join(chartsDir, filename)
			
			if err := os.WriteFile(filePath, []byte(svgContent), 0644); err != nil {
				log.Printf("Failed to save chart to file: %v", err)
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("failed to save chart: %v", err),
				}, nil
			}
			
			log.Printf("Chart saved to: %s", filePath)

			// Get server port from environment
			port := os.Getenv("PORT")
			if port == "" {
				port = "8080"
			}
			
			// Create HTTP URL for the chart
			chartURL := fmt.Sprintf("http://localhost:%s/charts/%s", port, filename)

			result := map[string]interface{}{
				"chart_type":   "line",
				"data_type":    "balance_trend",
				"image_url":    chartURL,
				"file_path":    filePath,
				"total_points": len(transactions),
				"message":      fmt.Sprintf("Generated balance trend chart with %d data points. View at: %s", len(transactions), chartURL),
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

// ============================================================================
// HACKATHON IDEAS
// ============================================================================
// Here are some ideas for custom tools you could build:
//
// 1. SAVINGS GOAL TRACKER
//    - Track progress toward savings goals
//    - Calculate how long until goal is reached
//    - Suggest optimal deposit amounts
//
// 2. BUDGET ANALYZER
//    - Set spending limits by category
//    - Alert when approaching limits
//    - Compare actual vs. planned spending
//
// 3. RECURRING PAYMENT DETECTOR
//    - Identify subscription payments
//    - Warn about upcoming bills
//    - Suggest savings opportunities
//
// 4. CASH FLOW FORECASTER
//    - Predict future balance based on patterns
//    - Identify potential low balance periods
//    - Suggest when to save vs. spend
//
// 5. SMART SAVINGS ADVISOR
//    - Analyze spare cash available
//    - Recommend savings deposits
//    - Calculate interest projections
//
// 6. SPENDING INSIGHTS
//    - Categorize spending automatically
//    - Compare to typical user patterns
//    - Highlight unusual activity
//
// 7. FINANCIAL HEALTH SCORE
//    - Calculate overall financial wellness
//    - Track improvements over time
//    - Provide actionable recommendations
//
// 8. PEER COMPARISON (anonymous)
//    - Compare savings rate to anonymized peers
//    - Show percentile rankings
//    - Motivate better habits
//
// 9. TAX ESTIMATION
//    - Track potential tax obligations
//    - Suggest amounts to set aside
//    - Generate tax reports
//
// 10. EMERGENCY FUND BUILDER
//     - Calculate needed emergency fund size
//     - Track progress toward goal
//     - Suggest automated savings plan
//
// ============================================================================

// ============================================================================
// LANGGRAPH-STYLE WORKFLOW SYSTEM
// ============================================================================
// Stateful workflow graph for building complex agent workflows

// Node represents a single step in the agent workflow graph
type Node struct {
	Name    string
	Handler func(ctx context.Context, state *GraphState) error
}

// GraphState holds the current state as the graph executes
type GraphState struct {
	Messages     []string
	CurrentTool  string
	ToolResult   interface{}
	UserID       string
	Conversation map[string]interface{}
	Error        error
}

// Graph represents a stateful workflow for the agent
type Graph struct {
	Nodes       map[string]*Node
	Edges       map[string][]string // node -> next possible nodes
	StartNode   string
	CurrentNode string
	State       *GraphState
}

// NewGraph creates a new agent workflow graph
func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[string]*Node),
		Edges: make(map[string][]string),
		State: &GraphState{
			Messages:     []string{},
			Conversation: make(map[string]interface{}),
		},
	}
}

// AddNode adds a node to the graph
func (g *Graph) AddNode(name string, handler func(ctx context.Context, state *GraphState) error) {
	g.Nodes[name] = &Node{
		Name:    name,
		Handler: handler,
	}
}

// AddEdge adds a directed edge from one node to another
func (g *Graph) AddEdge(from, to string) {
	g.Edges[from] = append(g.Edges[from], to)
}

// SetStart sets the starting node
func (g *Graph) SetStart(nodeName string) {
	g.StartNode = nodeName
	g.CurrentNode = nodeName
}

// Execute runs the graph starting from the start node
func (g *Graph) Execute(ctx context.Context, userID string) error {
	g.State.UserID = userID
	g.CurrentNode = g.StartNode

	visited := make(map[string]bool)

	for g.CurrentNode != "" {
		// Prevent infinite loops
		if visited[g.CurrentNode] {
			return fmt.Errorf("cycle detected at node: %s", g.CurrentNode)
		}
		visited[g.CurrentNode] = true

		node, exists := g.Nodes[g.CurrentNode]
		if !exists {
			return fmt.Errorf("node not found: %s", g.CurrentNode)
		}

		log.Printf("Executing node: %s", node.Name)

		// Execute the node handler
		if err := node.Handler(ctx, g.State); err != nil {
			g.State.Error = err
			return err
		}

		// Move to next node with conditional routing
		nextNodes := g.Edges[g.CurrentNode]
		if len(nextNodes) == 0 {
			// End of graph
			break
		}

		// Conditional routing for financial_help node
		if g.CurrentNode == "financial_help" {
			// Route based on balance check
			if financialRoute, ok := g.State.Conversation["financial_route"].(string); ok {
				if financialRoute == "save" {
					g.CurrentNode = "financial_save"
				} else {
					g.CurrentNode = "financial_help_low_funds"
				}
				continue
			}
		}

		// Conditional routing for financial_save node
		// Check if user chose chunks/DCA option (wants calendar reminders)
		if g.CurrentNode == "financial_save" {
			// Check if user indicated they want chunks/periodic investing
			if wantsReminders, ok := g.State.Conversation["wants_calendar_reminders"].(bool); ok && wantsReminders {
				log.Println("📅 User chose chunks/DCA - routing to investment_reminder")
				g.CurrentNode = "investment_reminder"
				continue
			} else {
				log.Println("✅ User chose lump sum or no response yet - skipping reminder node")
				// Skip to execute_tool
				g.CurrentNode = "execute_tool"
				continue
			}
		}

		// Conditional routing for orchestrator
		if g.CurrentNode == "orchestrator" {
			if route, ok := g.State.Conversation["route"].(string); ok {
				g.CurrentNode = route
				continue
			}
		}

		// Default: take the first edge
		g.CurrentNode = nextNodes[0]
	}

	return nil
}

// ConditionalRouter determines which node to execute next based on state
func (g *Graph) ConditionalRouter(condition func(state *GraphState) string) {
	// This would be used to route to different nodes based on conditions
	// Example: if balance < 0, go to "overdraft_warning" node
	// else go to "normal_response" node
}

// CreateFinancialAgentGraph creates an example financial agent workflow graph
func CreateFinancialAgentGraph(liminalExecutor core.ToolExecutor) *Graph {
	graph := NewGraph()

	// Root Node: Orchestrator - coordinates the entire workflow
	graph.AddNode("orchestrator", func(ctx context.Context, state *GraphState) error {
		log.Println("Orchestrator: Analyzing request and routing to appropriate handler...")
		// The orchestrator analyzes the user's request and routes to the correct first-layer node
		// In a real implementation, this would use Claude to classify the request type
		
		// Example routing logic (would be replaced with actual classification)
		userInput := state.Conversation["user_input"].(string)
		
		if strings.Contains(userInput, "deposit") || strings.Contains(userInput, "save") && !strings.Contains(userInput, "help") || strings.Contains(userInput, "put money") || strings.Contains(userInput, "add to savings") {
			state.Conversation["route"] = "deposit"
		} else if strings.Contains(userInput, "withdraw") || strings.Contains(userInput, "take out") || strings.Contains(userInput, "pull from savings") {
			state.Conversation["route"] = "withdraw"
		} else if strings.Contains(userInput, "image") || strings.Contains(userInput, "receipt") || strings.Contains(userInput, "split") {
			state.Conversation["route"] = "image_payment"
			log.Println("💙💙💙💙💙💙💙💙💙💙💙💙💙")
		} else if strings.Contains(userInput, "broke") || strings.Contains(userInput, "help") || 
			strings.Contains(userInput, "stats") || strings.Contains(userInput, "investing") || 
			strings.Contains(userInput, "saving") {
			state.Conversation["route"] = "financial_help"
			log.Println("🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴")
			log.Println("🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴")
			log.Println("🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴")
		} else {
			state.Conversation["route"] = "general_inquiry"
			log.Println("🍏🍏🍏🍏🍏🍏🍏🍏🍏🍏🍏🍏🍏")
		}
		
		return nil
	})

	// FIRST LAYER NODES - Three specialized handlers

	// Node 1: General Inquiry - handles standard queries using model's default functions
	graph.AddNode("general_inquiry", func(ctx context.Context, state *GraphState) error {
		log.Println("General Inquiry: Processing standard request...")
		// Use Claude's default capabilities for general questions
		// Check balance, view transactions, search users, etc.
		state.Conversation["handler_type"] = "general"
		state.Conversation["intent"] = "standard_query"
		
		// Check if user is asking for a chart/graph/balance trend
		userInput := state.Conversation["user_input"].(string)
		lowerInput := strings.ToLower(userInput)
		
		if strings.Contains(lowerInput, "chart") || strings.Contains(lowerInput, "graph") || 
			strings.Contains(lowerInput, "visualize") || strings.Contains(lowerInput, "plot") ||
			strings.Contains(lowerInput, "trend") || strings.Contains(lowerInput, "balance") && (strings.Contains(lowerInput, "show") || strings.Contains(lowerInput, "see")) {
			log.Println("📊 Detected balance trend chart request")
			state.Conversation["chart_requested"] = true
		}
		
		return nil
	})

	// Node 2: Image Payment - handles receipt splitting and image-based payments
	graph.AddNode("image_payment", func(ctx context.Context, state *GraphState) error {
		log.Println("Image Payment: Processing receipt/image for payment splitting...")
		// Use Claude vision to analyze receipt images
		// Extract amounts, calculate splits, identify friends to pay
		state.Conversation["handler_type"] = "image_payment"
		state.CurrentTool = "process_receipt_image"
		
		// Example: Parse receipt and determine split amounts
		state.ToolResult = map[string]interface{}{
			"total_amount": 50.00,
			"split_count":  2,
			"per_person":   25.00,
			"currency":     "USD",
			"recipients":   []string{"@alice", "@bob"},
		}
		return nil
	})

	// Node 3: Financial Help (Get Stats) - provides financial insights and recommendations
	// This node checks balance and routes to either "save" or "help" based on funds
	graph.AddNode("financial_help", func(ctx context.Context, state *GraphState) error {
		log.Println("Financial Help: Analyzing financial situation and checking balance...")
		
		// Send intermediate message to user
		state.Messages = append(state.Messages, "💰 Step 1/3: Checking your account balance...")
		
		// Fetch user's balance to determine routing
		balanceResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
			UserID:    state.UserID,
			Tool:      "get_balance",
			Input:     []byte("{}"),
			RequestID: fmt.Sprintf("balance-check-%d", time.Now().Unix()),
		})
		
		if err != nil {
			log.Printf("Failed to fetch balance: %v", err)
			// Default to help route if balance fetch fails
			state.Conversation["financial_route"] = "help"
			state.Conversation["handler_type"] = "financial_help"
			return nil
		}
		
		// Parse balance data
		var balanceData map[string]interface{}
		var usdcBalance, eurcBalance, lilBalance float64
		
		if balanceResponse.Success {
			if err := json.Unmarshal(balanceResponse.Data, &balanceData); err == nil {
				// Parse wallet balances
				if wallet, ok := balanceData["wallet"].(map[string]interface{}); ok {
					if balances, ok := wallet["balances"].([]interface{}); ok {
						for _, bal := range balances {
							if balMap, ok := bal.(map[string]interface{}); ok {
								currency, _ := balMap["currency"].(string)
								amountStr, _ := balMap["amount"].(string)
								amount := 0.0
								fmt.Sscanf(amountStr, "%f", &amount)
								
								switch currency {
								case "USDC":
									usdcBalance = amount
								case "EURC":
									eurcBalance = amount
								case "LIL":
									lilBalance = amount
								}
							}
						}
					}
				}
			}
		}
		
		log.Printf("💰 Balance check: USDC=%.2f, EURC=%.2f, LIL=%.2f", usdcBalance, eurcBalance, lilBalance)
		
		// Send balance summary to user
		balanceSummary := fmt.Sprintf("✅ Balance found: %.2f USDC, %.2f EURC, %.2f LIL", usdcBalance, eurcBalance, lilBalance)
		state.Messages = append(state.Messages, balanceSummary)
		
		// Determine route based on balance
		// Save route: > 2 USDC/EURC AND > 5 LIL
		// Help route: less than those amounts
		hasEnoughStablecoin := usdcBalance > 2 || eurcBalance > 2
		hasEnoughLIL := lilBalance > 5
		
		if hasEnoughStablecoin && hasEnoughLIL {
			log.Println("✅ User has sufficient funds → SAVE route")
			state.Conversation["financial_route"] = "save"
			state.Conversation["handler_type"] = "financial_save"
		} else {
			log.Println("⚠️ User has low funds → HELP route")
			state.Conversation["financial_route"] = "help"
			state.Conversation["handler_type"] = "financial_help"
		}
		
		// Store balance info for later use
		state.Conversation["balance_usdc"] = usdcBalance
		state.Conversation["balance_eurc"] = eurcBalance
		state.Conversation["balance_lil"] = lilBalance
		
		return nil
	})
	
	// Node 3a: Save Route - for users with sufficient funds
	graph.AddNode("financial_save", func(ctx context.Context, state *GraphState) error {
		log.Println("Financial Save Route: User has sufficient funds for saving/investing...")
		state.Conversation["handler_type"] = "financial_save"
		state.CurrentTool = "analyze_savings_opportunities"
		
		// Check if this is a follow-up message where user chose chunks/DCA
		if len(state.Messages) > 0 {
			lastMessage := strings.ToLower(state.Messages[len(state.Messages)-1])
			// Detect if user wants chunks/periodic/DCA investing
			if strings.Contains(lastMessage, "chunk") || 
			   strings.Contains(lastMessage, "period") ||
			   strings.Contains(lastMessage, "week") ||
			   strings.Contains(lastMessage, "month") ||
			   strings.Contains(lastMessage, "spread") ||
			   strings.Contains(lastMessage, "split") ||
			   strings.Contains(lastMessage, "reminder") ||
			   strings.Contains(lastMessage, "2") || // Option 2
			   strings.Contains(lastMessage, "dca") {
				log.Println("📅 Detected user wants chunks/periodic investing - will offer reminders")
				state.Conversation["wants_calendar_reminders"] = true
			}
		}
		
		// Send intermediate message
		state.Messages = append(state.Messages, "📊 Step 2/3: Comparing vault rates to find your best option...")
		
		// Get user's balance from previous node
		usdcBalance, _ := state.Conversation["balance_usdc"].(float64)
		eurcBalance, _ := state.Conversation["balance_eurc"].(float64)
		
		// Fetch vault rates to compare options
		vaultRatesResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
			UserID:    state.UserID,
			Tool:      "get_vault_rates",
			Input:     []byte("{}"),
			RequestID: fmt.Sprintf("vault-rates-%d", time.Now().Unix()),
		})
		
		var usdcAPY, eurcAPY float64
		var bestCurrency string
		var bestAPY float64
		
		if err == nil && vaultRatesResponse.Success {
			var ratesData map[string]interface{}
			if err := json.Unmarshal(vaultRatesResponse.Data, &ratesData); err == nil {
				if vaults, ok := ratesData["vaults"].([]interface{}); ok {
					for _, vault := range vaults {
						if vaultMap, ok := vault.(map[string]interface{}); ok {
							currency, _ := vaultMap["currency"].(string)
							apyStr, _ := vaultMap["apy"].(string)
							apy := 0.0
							fmt.Sscanf(apyStr, "%f", &apy)
							
							if currency == "USDC" {
								usdcAPY = apy
							} else if currency == "EURC" {
								eurcAPY = apy
							}
						}
					}
				}
			}
		}
		
		log.Printf("💰 Vault rates: USDC=%.2f%%, EURC=%.2f%%", usdcAPY, eurcAPY)
		
		// Compare rates for currencies user actually has
		if usdcBalance > 2 && eurcBalance > 2 {
			// User has both - recommend the higher rate
			if usdcAPY >= eurcAPY {
				bestCurrency = "USDC"
				bestAPY = usdcAPY
			} else {
				bestCurrency = "EURC"
				bestAPY = eurcAPY
			}
		} else if usdcBalance > 2 {
			// Only has USDC
			bestCurrency = "USDC"
			bestAPY = usdcAPY
		} else if eurcBalance > 2 {
			// Only has EURC
			bestCurrency = "EURC"
			bestAPY = eurcAPY
		} else {
			// Fallback (shouldn't happen but just in case)
			bestCurrency = "USDC"
			bestAPY = usdcAPY
		}
		
		log.Printf("✅ Best savings option: %s at %.2f%% APY", bestCurrency, bestAPY)
		
		// Send rate comparison result
		rateMessage := fmt.Sprintf("✅ Best rate: %s vault at %.2f%% APY", bestCurrency, bestAPY)
		state.Messages = append(state.Messages, rateMessage)
		
		// Calculate available amount to save (keep some for spending)
		var availableToSave float64
		if bestCurrency == "USDC" {
			availableToSave = usdcBalance - 2 // Keep minimum balance
		} else {
			availableToSave = eurcBalance - 2
		}
		
		// Generate personalized saving/investing recommendations
		recommendations := []string{
			fmt.Sprintf("💰 Best option: Save in %s vault earning %.2f%% APY", bestCurrency, bestAPY),
			fmt.Sprintf("💵 You have %.2f %s available to save", availableToSave, bestCurrency),
		}
		
		// Add comparison if user has both currencies
		if usdcBalance > 2 && eurcBalance > 2 {
			recommendations = append(recommendations, 
				fmt.Sprintf("📊 Rate comparison: USDC %.2f%% vs EURC %.2f%%", usdcAPY, eurcAPY))
		}
		
		// Generate comparison chart: Lump Sum vs Dollar-Cost Averaging
		state.Messages = append(state.Messages, "📈 Step 3/3: Creating investment strategy comparison...")
		chartSVG := generateInvestmentComparisonChart(availableToSave, bestAPY, bestCurrency)
		
		// Save chart to file
		chartsDir := filepath.Join(".", "charts")
		os.MkdirAll(chartsDir, 0755)
		chartFileName := fmt.Sprintf("investment-comparison-%d.svg", time.Now().Unix())
		chartFilePath := filepath.Join(chartsDir, chartFileName)
		os.WriteFile(chartFilePath, []byte(chartSVG), 0644)
		chartURL := fmt.Sprintf("http://localhost:8080/charts/%s", chartFileName)
		
		recommendations = append(recommendations, []string{
			"",
			"📊 Investment Growth Comparison:",
			fmt.Sprintf("![Investment Comparison](%s)", chartURL),
			"",
		}...)
		
		// Educational content about investment strategies
		recommendations = append(recommendations, []string{
			"",
			"💡 Investment Strategy Options:",
			"",
			"🎯 Lump Sum (All at once):",
			"   • Pro: Start earning interest immediately on full amount",
			"   • Pro: Simpler - one transaction and done",
			"   • Pro: Better if rates are expected to drop",
			"   • Con: Higher risk if market/rates fluctuate",
			"",
			"📅 Dollar-Cost Averaging (Chunks over time):",
			"   • Pro: Reduces timing risk - spreads deposits over weeks/months",
			"   • Pro: Helps build a savings habit with regular deposits",
			"   • Pro: Less stressful - you don't have to pick the 'perfect' time",
			"   • Con: May earn less interest initially on uninvested funds",
			"",
			"🎓 Recommendation:",
			fmt.Sprintf("   With %.2f %s available:", availableToSave, bestCurrency),
			"   • Conservative: Deposit 50%% now, split rest over 4 weeks",
			"   • Moderate: Deposit 75%% now, rest next week",
			"   • Aggressive: Deposit all now to maximize APY immediately",
			"",
			"💪 Choose based on your comfort level and financial goals!",
		}...)
		
		recommendations = append(recommendations, []string{
			"",
			"❓ Which approach works best for you?",
			"   1️⃣  Lump sum - Deposit all at once",
			"   2️⃣  Chunks - Spread deposits over time (I can set up calendar reminders!)",
			"",
			"💡 Ready to start saving now?",
			fmt.Sprintf("I can help you deposit into your %s savings vault (%.2f%% APY) right away!", bestCurrency, bestAPY),
			fmt.Sprintf("Just say \"deposit [amount] %s\" to get started.", bestCurrency),
		}...)
		
		state.ToolResult = map[string]interface{}{
			"status":              "sufficient_funds",
			"recommendation_type": "savings_investment",
			"best_currency":       bestCurrency,
			"best_apy":            bestAPY,
			"usdc_apy":            usdcAPY,
			"eurc_apy":            eurcAPY,
			"available_to_save":   availableToSave,
			"recommendations":     recommendations,
			"can_deposit":         true,
		}
		state.Conversation["can_deposit"] = true
		state.Conversation["suggested_currency"] = bestCurrency
		return nil
	})
	
	// Node 3c: Investment Reminder - asks about setting up periodic investment reminders
	// This node should only be reached if user chose chunks/DCA option
	graph.AddNode("investment_reminder", func(ctx context.Context, state *GraphState) error {
		log.Println("Investment Reminder: User chose chunks/DCA - offering calendar reminders...")
		
		// Get the recommended currency and amount
		bestCurrency := "USDC"
		availableToSave := 0.0
		
		if result, ok := state.ToolResult.(map[string]interface{}); ok {
			if curr, ok := result["best_currency"].(string); ok {
				bestCurrency = curr
			}
			if available, ok := result["available_to_save"].(float64); ok {
				availableToSave = available
			}
		}
		
		// Store reminder context for Claude to use
		// Suggest the tool that Claude should call when user responds
		state.Conversation["reminder_context"] = map[string]interface{}{
			"currency":          bestCurrency,
			"available_amount":  availableToSave,
			"reminder_type":     "investment_schedule",
			"frequency_options": []string{"weekly", "bi-weekly", "monthly"},
			"suggested_tool":    "create_calendar_reminder",
		}
		state.Conversation["next_action"] = "If user chooses a frequency, use create_calendar_reminder tool with their chosen frequency"
		
		// Calculate suggested amounts for different frequencies
		weeklyAmount := availableToSave / 4   // Spread over 4 weeks
		biWeeklyAmount := availableToSave / 2 // Spread over 2 periods
		monthlyAmount := availableToSave / 3  // Spread over 3 months
		
		reminderInfo := map[string]interface{}{
			"status":           "reminder_offered",
			"message_to_user":  "Would you like me to help you set up calendar reminders for your periodic investments?",
			"available_amount": availableToSave,
			"best_currency":    bestCurrency,
			"frequency_options": map[string]interface{}{
				"weekly": map[string]interface{}{
					"description": "Invest every week",
					"amount":      weeklyAmount,
					"duration":    "4 weeks",
				},
				"bi-weekly": map[string]interface{}{
					"description": "Invest every 2 weeks",
					"amount":      biWeeklyAmount,
					"duration":    "1 month",
				},
				"monthly": map[string]interface{}{
					"description": "Invest once a month",
					"amount":      monthlyAmount,
					"duration":    "3 months",
				},
			},
			"benefits": []string{
				"📅 Never miss your investment schedule",
				"💪 Build consistent savings habits",
				"🔔 Get notified when it's time to deposit",
				"📈 Stay on track with your financial goals",
			},
			"deposit_now_prompt": fmt.Sprintf("💰 Would you also like to make your first deposit now?\n\nYou have %.2f %s available. I can help you deposit any amount into savings right away!", availableToSave, bestCurrency),
		}
		
		// Store in tool result for Claude to present
		state.ToolResult = reminderInfo
		state.Conversation["handler_type"] = "investment_reminder"
		state.Conversation["can_deposit"] = true
		state.Conversation["deposit_currency"] = bestCurrency
		
		return nil
	})
	
	// Node 3b: Help Route - for users with low funds
	graph.AddNode("financial_help_low_funds", func(ctx context.Context, state *GraphState) error {
		log.Println("Financial Help Route: User needs assistance with low funds...")
		state.Conversation["handler_type"] = "financial_help_low_funds"
		state.CurrentTool = "analyze_financial_assistance"
		
		// Send intermediate message
		state.Messages = append(state.Messages, "🔍 Analyzing your spending patterns to find ways to save...")
		
		// Get balance for income estimation
		usdcBalance, _ := state.Conversation["balance_usdc"].(float64)
		eurcBalance, _ := state.Conversation["balance_eurc"].(float64)
		totalBalance := usdcBalance + eurcBalance
		
		// Store for next node
		state.Conversation["total_balance"] = totalBalance
		
		// Generate initial help recommendations
		state.ToolResult = map[string]interface{}{
			"status":            "low_funds_analysis_needed",
			"recommendation_type": "financial_assistance",
			"next_step": "spending_analysis",
		}
		return nil
	})
	
	// Node 3b-2: Spending Analysis - flags unnecessary and excessive expenses
	graph.AddNode("spending_analysis", func(ctx context.Context, state *GraphState) error {
		log.Println("Spending Analysis: Flagging unnecessary and excessive expenses...")
		state.Conversation["handler_type"] = "spending_analysis"
		
		// Fetch transactions
		txResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
			UserID:    state.UserID,
			Tool:      "get_transactions",
			Input:     []byte(`{"limit": 50}`),
			RequestID: fmt.Sprintf("txs-analysis-%d", time.Now().Unix()),
		})
		
		if err != nil || !txResponse.Success {
			log.Printf("Failed to fetch transactions: %v", err)
			state.ToolResult = map[string]interface{}{
				"status": "error",
				"message": "Could not analyze spending",
			}
			return nil
		}
		
		// Parse transactions
		var txData map[string]interface{}
		json.Unmarshal(txResponse.Data, &txData)
		transactions, _ := txData["transactions"].([]interface{})
		
		// Get user's balance (proxy for income level)
		totalBalance, _ := state.Conversation["total_balance"].(float64)
		estimatedMonthlyIncome := totalBalance * 2 // Rough estimate
		
		// Categorize and flag spending
		flaggedSpending := analyzeAndFlagSpending(transactions, estimatedMonthlyIncome)
		
		// Generate visualization
		chartSVG := generateFlaggedSpendingChart(flaggedSpending)
		
		// Save chart
		chartsDir := filepath.Join(".", "charts")
		os.MkdirAll(chartsDir, 0755)
		chartFileName := fmt.Sprintf("flagged-spending-%d.svg", time.Now().Unix())
		chartFilePath := filepath.Join(chartsDir, chartFileName)
		os.WriteFile(chartFilePath, []byte(chartSVG), 0644)
		chartURL := fmt.Sprintf("http://localhost:8080/charts/%s", chartFileName)
		
		// Build recommendations
		recommendations := []string{
			"📊 Your Spending Analysis:",
			fmt.Sprintf("![Flagged Spending](%s)", chartURL),
			"",
		}
		
		if len(flaggedSpending.Unnecessary) > 0 {
			recommendations = append(recommendations, "🚩 Unnecessary Expenses:")
			for _, item := range flaggedSpending.Unnecessary {
				recommendations = append(recommendations, fmt.Sprintf("   • %s: $%.2f - %s", item.Category, item.Amount, item.Reason))
			}
			recommendations = append(recommendations, "")
		}
		
		if len(flaggedSpending.Excessive) > 0 {
			recommendations = append(recommendations, "⚠️ Excessive Essential Expenses:")
			for _, item := range flaggedSpending.Excessive {
				recommendations = append(recommendations, fmt.Sprintf("   • %s: $%.2f - %s", item.Category, item.Amount, item.Reason))
			}
			recommendations = append(recommendations, "")
		}
		
		totalSavings := flaggedSpending.TotalUnnecessary + flaggedSpending.TotalExcessive
		if totalSavings > 0 {
			recommendations = append(recommendations, []string{
				fmt.Sprintf("💡 Potential Monthly Savings: $%.2f", totalSavings),
				"",
				"📌 Action Steps:",
				"1. Cancel unused subscriptions immediately",
				"2. Set alerts for essential expense categories",
				"3. Consider cheaper alternatives for flagged items",
				"4. Use weekly spending goals to track progress",
			}...)
		} else {
			recommendations = append(recommendations, "✅ Your spending looks reasonable! Focus on increasing income.")
		}
		
		state.ToolResult = map[string]interface{}{
			"status":              "analysis_complete",
			"recommendation_type": "spending_flags",
			"recommendations":     recommendations,
			"flagged_spending":    flaggedSpending,
			"potential_savings":   totalSavings,
		}
		state.Conversation["flagged_spending"] = flaggedSpending
		state.Conversation["potential_savings"] = totalSavings
		state.Conversation["handler_type"] = "budget_recommendations"
		state.Conversation["next_step"] = "budget_recommendations"
		return nil
	})

	// Budget recommendations - explain WHY spending is bad and set weekly budget goals
	graph.AddNode("budget_recommendations", func(ctx context.Context, state *GraphState) error {
		log.Println("Creating personalized budget recommendations...")
		
		flaggedSpending := state.Conversation["flagged_spending"].(FlaggedSpending)
		potentialSavings := state.Conversation["potential_savings"].(float64)
		totalBalance := state.Conversation["total_balance"].(float64)
		estimatedMonthlyIncome := totalBalance * 2
		
		var recommendations []string
		recommendations = append(recommendations, "💭 Let me explain what I found and how we can fix this:\n")
		
		// Explain WHY each flagged category is problematic
		if len(flaggedSpending.Unnecessary) > 0 {
			recommendations = append(recommendations, "🚫 Unnecessary Expenses - Why These Matter:")
			for _, item := range flaggedSpending.Unnecessary {
				var explanation string
				switch item.Category {
				case "entertainment":
					percent := (item.Amount / estimatedMonthlyIncome) * 100
					explanation = fmt.Sprintf("Entertainment (%.0f%% of income) - Financial experts recommend keeping entertainment under 5-10%% of your income. This helps ensure you're prioritizing savings and essentials first.", percent)
				case "subscription":
					explanation = fmt.Sprintf("Subscriptions ($%.2f/month) - Many people pay for services they rarely use. Audit your subscriptions quarterly - even small recurring charges add up to hundreds per year.", item.Amount)
				default:
					explanation = fmt.Sprintf("%s spending - While not essential, cutting this can free up cash for emergencies or savings goals.", item.Category)
				}
				recommendations = append(recommendations, fmt.Sprintf("   • %s: $%.2f\n     💡 %s", item.Category, item.Amount, explanation))
			}
			recommendations = append(recommendations, "")
		}
		
		if len(flaggedSpending.Excessive) > 0 {
			recommendations = append(recommendations, "⚠️  Excessive Essential Expenses - Why These Are Too High:")
			for _, item := range flaggedSpending.Excessive {
				recommendations = append(recommendations, fmt.Sprintf("   • %s: $%.2f\n     💡 %s", item.Category, item.Amount, item.Reason))
			}
			recommendations = append(recommendations, "")
		}
		
		// Specific recommendations for each flagged category
		recommendations = append(recommendations, "✨ Personalized Recommendations:")
		for _, item := range flaggedSpending.Unnecessary {
			var rec string
			switch item.Category {
			case "entertainment":
				rec = "→ Entertainment: Try free alternatives - public parks, library events, free museum days. Set a monthly entertainment budget of $" + fmt.Sprintf("%.0f", estimatedMonthlyIncome*0.05)
			case "subscription":
				rec = "→ Subscriptions: Cancel services you haven't used in 30 days. Share family plans to split costs. Consider rotating subscriptions monthly."
			default:
				rec = fmt.Sprintf("→ %s: Pause spending here for 30 days and see if you miss it. If not, cut permanently.", item.Category)
			}
			recommendations = append(recommendations, "   "+rec)
		}
		
		for _, item := range flaggedSpending.Excessive {
			var rec string
				switch item.Category {
			case "food":
				target := estimatedMonthlyIncome * 0.15
				rec = fmt.Sprintf("→ Food: Meal prep on Sundays, buy generic brands, use grocery apps for discounts. Target: $%.0f/month (currently $%.0f)", target, item.Amount)
			case "travel":
				target := estimatedMonthlyIncome * 0.10
				rec = fmt.Sprintf("→ Travel: Use public transit, carpool apps, or bike when possible. Consider a monthly transit pass. Target: $%.0f/month (currently $%.0f)", target, item.Amount)
			case "electronics":
				rec = "→ Electronics: Buy refurbished, wait for sales, or use buy-nothing groups. Electronics are rarely urgent purchases."
			default:
				rec = fmt.Sprintf("→ %s: Research cheaper alternatives or negotiate better rates with providers.", item.Category)
			}
			recommendations = append(recommendations, "   "+rec)
		}
		recommendations = append(recommendations, "")
		
		// Calculate weekly budget goals
		weeklySavingsGoal := potentialSavings / 4.0
		currentWeeklySpending := (flaggedSpending.TotalUnnecessary + flaggedSpending.TotalExcessive) / 4.0
		targetWeeklySpending := currentWeeklySpending - weeklySavingsGoal
		
		// Calculate recommended weekly budgets by category
		weeklyIncome := estimatedMonthlyIncome / 4.0
		weeklyBudget := map[string]float64{
			"Essentials (rent, utilities, insurance)": weeklyIncome * 0.50,
			"Food & groceries":                        weeklyIncome * 0.15,
			"Transportation":                          weeklyIncome * 0.10,
			"Savings":                                 weeklyIncome * 0.20,
			"Personal & entertainment":                weeklyIncome * 0.05,
		}
		
		recommendations = append(recommendations, []string{
			"📊 Your Weekly Budget Goal:",
			fmt.Sprintf("Current weekly spending: $%.2f", currentWeeklySpending),
			fmt.Sprintf("Target weekly spending: $%.2f", targetWeeklySpending),
			fmt.Sprintf("Weekly savings goal: $%.2f", weeklySavingsGoal),
			"",
			"📋 Recommended Weekly Budget Breakdown:",
		}...)
		
		for category, amount := range weeklyBudget {
			recommendations = append(recommendations, fmt.Sprintf("   • %s: $%.2f", category, amount))
		}
		
		recommendations = append(recommendations, []string{
			"",
			"🎯 How to Track Your Weekly Budget:",
			"   1. Set a weekly spending limit alert on your phone",
			"   2. Check your balance every Monday morning",
			"   3. If you're over budget mid-week, pause non-essential spending",
			"   4. Celebrate when you hit your weekly savings goal!",
			"",
			fmt.Sprintf("💰 If you follow this plan, you could save $%.2f per month ($%.2f per year)!", potentialSavings, potentialSavings*12),
			"",
			"💡 Once you cut these expenses, would you like to start building savings?",
			"I can help you set up automatic deposits into high-yield savings vaults.",
			"Just let me know when you're ready to start saving!",
		}...)
		
		state.ToolResult = map[string]interface{}{
			"status":                  "recommendations_ready",
			"recommendation_type":     "budget_plan",
			"recommendations":         recommendations,
			"weekly_savings_goal":     weeklySavingsGoal,
			"target_weekly_spending":  targetWeeklySpending,
			"weekly_budget_breakdown": weeklyBudget,
			"can_start_saving":        true,
		}
		return nil
	})

	// Withdraw - educational withdrawal with safety checks
	graph.AddNode("withdraw", func(ctx context.Context, state *GraphState) error {
		log.Println("Withdraw: Analyzing withdrawal safety and providing education...")
		
		// Fetch current balances
		balanceRequest := map[string]interface{}{}
		balanceRequestJSON, _ := json.Marshal(balanceRequest)
		
		balanceResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
			UserID:    state.Conversation["user_id"].(string),
			Tool:      "get_balance",
			Input:     balanceRequestJSON,
			RequestID: state.Conversation["request_id"].(string),
		})
		
		if err != nil || !balanceResponse.Success {
			log.Printf("Failed to fetch balance: %v", err)
			return nil
		}
		
		savingsResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
			UserID:    state.Conversation["user_id"].(string),
			Tool:      "get_savings_balance",
			Input:     balanceRequestJSON,
			RequestID: state.Conversation["request_id"].(string),
		})
		
		if err != nil || !savingsResponse.Success {
			log.Printf("Failed to fetch savings balance: %v", err)
			return nil
		}
		
		// Parse balances
		var walletBalance, savingsBalance float64
		var walletData map[string]interface{}
		if err := json.Unmarshal(balanceResponse.Data, &walletData); err == nil {
			if balances, ok := walletData["balances"].([]interface{}); ok {
				for _, bal := range balances {
					if balMap, ok := bal.(map[string]interface{}); ok {
						if currency, ok := balMap["currency"].(string); ok && (currency == "USD" || currency == "EUR") {
							if balStr, ok := balMap["balance"].(string); ok {
								var amount float64
								fmt.Sscanf(balStr, "%f", &amount)
								walletBalance += amount
							}
						}
					}
				}
			}
		}
		
		var savingsData map[string]interface{}
		if err := json.Unmarshal(savingsResponse.Data, &savingsData); err == nil {
			if positions, ok := savingsData["positions"].([]interface{}); ok {
				for _, pos := range positions {
					if posMap, ok := pos.(map[string]interface{}); ok {
						if balStr, ok := posMap["balance"].(string); ok {
							var amount float64
							fmt.Sscanf(balStr, "%f", &amount)
							savingsBalance += amount
						}
					}
				}
			}
		}
		
		// Extract withdrawal details from user message (Claude should extract this)
		userInput := state.Conversation["user_input"].(string)
		log.Printf("User withdrawal request: %s", userInput)
		
		totalLiquidity := walletBalance + savingsBalance
		liquidityRatio := walletBalance / totalLiquidity
		
		// Determine withdrawal safety
		// Unsafe: Low wallet balance (<20% of total) and withdrawing from savings
		// Safe: Sufficient wallet balance (>=20% of total) for big purchase
		isUnsafeWithdrawal := liquidityRatio < 0.20 && savingsBalance > 0
		
		var recommendations []string
		
		if isUnsafeWithdrawal {
			// UNSAFE WITHDRAWAL - Very little liquidity, pulling from savings
			recommendations = append(recommendations, []string{
				"⚠️ Withdrawal Safety Check:",
				"",
				fmt.Sprintf("Your current situation:"),
				fmt.Sprintf("   • Wallet (liquid): $%.2f (%.0f%% of total)", walletBalance, liquidityRatio*100),
				fmt.Sprintf("   • Savings: $%.2f", savingsBalance),
				fmt.Sprintf("   • Total: $%.2f", totalLiquidity),
				"",
				"🚨 This is a hasty withdrawal situation:",
				"",
				"Why this matters:",
				"   1. You have very little liquid cash available (less than 20% of your total)",
				"   2. You're pulling from your savings that's earning interest",
				"   3. This could become a habit that prevents wealth building",
				"",
				"💡 What you should know:",
				"   • Financial experts recommend keeping 3-6 months expenses liquid",
				"   • Savings should be for emergencies or planned goals, not daily spending",
				"   • Frequent withdrawals mean you're living above your means",
				"",
				"📚 Education - The Liquidity Trap:",
				"When you withdraw from savings for non-emergencies, you lose:",
				"   • Future compound interest earnings",
				"   • Emergency fund protection",
				"   • Financial flexibility for opportunities",
				"",
				fmt.Sprintf("Example: If you leave $%.2f in savings at 5%% APY, you'd earn $%.2f per year.", savingsBalance, savingsBalance*0.05),
				"By withdrawing, you're giving up this passive income.",
				"",
				"✅ I'll allow this withdrawal, BUT...",
				"",
				"To protect your financial health, I'm going to help you set a weekly spending budget.",
				"This will prevent you from needing emergency withdrawals in the future.",
			}...)
		} else {
			// SAFE WITHDRAWAL - Sufficient liquidity for big purchase
			recommendations = append(recommendations, []string{
				"✅ Withdrawal Safety Check:",
				"",
				fmt.Sprintf("Your current situation:"),
				fmt.Sprintf("   • Wallet (liquid): $%.2f (%.0f%% of total)", walletBalance, liquidityRatio*100),
				fmt.Sprintf("   • Savings: $%.2f", savingsBalance),
				fmt.Sprintf("   • Total: $%.2f", totalLiquidity),
				"",
				"✅ This is a safe withdrawal situation:",
				"You have sufficient liquid funds available, so withdrawing from savings is reasonable.",
				"",
				"💡 Quick Financial Education:",
				"",
				"Even though this is safe, here's what you should consider:",
				"   1. Opportunity Cost: Money in savings earns compound interest",
				"   2. Rebuilding: Plan to replenish your savings after this withdrawal",
				"   3. Goals: Make sure this purchase aligns with your financial priorities",
				"",
				"📊 Smart Withdrawal Practices:",
				"   • Only withdraw for planned big purchases or emergencies",
				"   • Try to maintain at least 50% of your wealth in savings/investments",
				"   • Set a goal to replace withdrawn funds within 3 months",
				"",
				fmt.Sprintf("💰 Cost of Withdrawal: At 5%% APY, every $100 withdrawn costs you $5/year in lost earnings."),
				"",
				"✅ You're cleared to proceed with this withdrawal.",
				"Just specify the amount and currency, and I'll help you withdraw.",
			}...)
		}
		
		// Calculate recommended weekly budget for unsafe withdrawals
		var weeklyBudgetLimit float64
		if isUnsafeWithdrawal {
			// Suggest conservative budget: 15% of total liquidity per week
			weeklyBudgetLimit = totalLiquidity * 0.15
			recommendations = append(recommendations, []string{
				"",
				"🎯 Recommended Weekly Budget:",
				fmt.Sprintf("   • Weekly spending limit: $%.2f", weeklyBudgetLimit),
				fmt.Sprintf("   • This keeps you at ~60%% of your income for essentials"),
				"   • Leaves room to rebuild your liquid funds",
				"",
				"Would you like me to set this weekly budget goal for you?",
				"(I can track your spending and alert you when you're approaching the limit)",
			}...)
		}
		
		state.ToolResult = map[string]interface{}{
			"status":                "withdrawal_analyzed",
			"recommendation_type":   "withdrawal_education",
			"recommendations":       recommendations,
			"is_unsafe_withdrawal": isUnsafeWithdrawal,
			"wallet_balance":        walletBalance,
			"savings_balance":       savingsBalance,
			"liquidity_ratio":       liquidityRatio,
			"weekly_budget_limit":   weeklyBudgetLimit,
			"can_withdraw":          true,
		}
		
		state.Conversation["handler_type"] = "withdraw"
		state.Conversation["is_unsafe_withdrawal"] = isUnsafeWithdrawal
		state.Conversation["weekly_budget_limit"] = weeklyBudgetLimit
		
		return nil
	})

	// Withdraw - educational withdrawal with safety checks
	graph.AddNode("withdraw", func(ctx context.Context, state *GraphState) error {
		log.Println("Withdraw: Analyzing withdrawal safety and providing education...")
		
		// Fetch current balances
		balanceRequest := map[string]interface{}{}
		balanceRequestJSON, _ := json.Marshal(balanceRequest)
		
		balanceResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
			UserID:    state.Conversation["user_id"].(string),
			Tool:      "get_balance",
			Input:     balanceRequestJSON,
			RequestID: state.Conversation["request_id"].(string),
		})
		
		if err != nil || !balanceResponse.Success {
			log.Printf("Failed to fetch balance: %v", err)
			return nil
		}
		
		savingsResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
			UserID:    state.Conversation["user_id"].(string),
			Tool:      "get_savings_balance",
			Input:     balanceRequestJSON,
			RequestID: state.Conversation["request_id"].(string),
		})
		
		if err != nil || !savingsResponse.Success {
			log.Printf("Failed to fetch savings balance: %v", err)
			return nil
		}
		
		// Parse balances
		var walletBalance, savingsBalance float64
		var walletData map[string]interface{}
		if err := json.Unmarshal(balanceResponse.Data, &walletData); err == nil {
			if balances, ok := walletData["balances"].([]interface{}); ok {
				for _, bal := range balances {
					if balMap, ok := bal.(map[string]interface{}); ok {
						if currency, ok := balMap["currency"].(string); ok && (currency == "USD" || currency == "EUR") {
							if balStr, ok := balMap["balance"].(string); ok {
								var amount float64
								fmt.Sscanf(balStr, "%f", &amount)
								walletBalance += amount
							}
						}
					}
				}
			}
		}
		
		var savingsData map[string]interface{}
		if err := json.Unmarshal(savingsResponse.Data, &savingsData); err == nil {
			if positions, ok := savingsData["positions"].([]interface{}); ok {
				for _, pos := range positions {
					if posMap, ok := pos.(map[string]interface{}); ok {
						if balStr, ok := posMap["balance"].(string); ok {
							var amount float64
							fmt.Sscanf(balStr, "%f", &amount)
							savingsBalance += amount
						}
					}
				}
			}
		}
		
		// Extract withdrawal details from user message (Claude should extract this)
		userInput := state.Conversation["user_input"].(string)
		log.Printf("User withdrawal request: %s", userInput)
		
		totalLiquidity := walletBalance + savingsBalance
		liquidityRatio := walletBalance / totalLiquidity
		
		// Determine withdrawal safety
		// Unsafe: Low wallet balance (<20% of total) and withdrawing from savings
		// Safe: Sufficient wallet balance (>=20% of total) for big purchase
		isUnsafeWithdrawal := liquidityRatio < 0.20 && savingsBalance > 0
		
		var recommendations []string
		
		if isUnsafeWithdrawal {
			// UNSAFE WITHDRAWAL - Very little liquidity, pulling from savings
			recommendations = append(recommendations, []string{
				"⚠️ Withdrawal Safety Check:",
				"",
				fmt.Sprintf("Your current situation:"),
				fmt.Sprintf("   • Wallet (liquid): $%.2f (%.0f%% of total)", walletBalance, liquidityRatio*100),
				fmt.Sprintf("   • Savings: $%.2f", savingsBalance),
				fmt.Sprintf("   • Total: $%.2f", totalLiquidity),
				"",
				"🚨 This is a hasty withdrawal situation:",
				"",
				"Why this matters:",
				"   1. You have very little liquid cash available (less than 20% of your total)",
				"   2. You're pulling from your savings that's earning interest",
				"   3. This could become a habit that prevents wealth building",
				"",
				"💡 What you should know:",
				"   • Financial experts recommend keeping 3-6 months expenses liquid",
				"   • Savings should be for emergencies or planned goals, not daily spending",
				"   • Frequent withdrawals mean you're living above your means",
				"",
				"📚 Education - The Liquidity Trap:",
				"When you withdraw from savings for non-emergencies, you lose:",
				"   • Future compound interest earnings",
				"   • Emergency fund protection",
				"   • Financial flexibility for opportunities",
				"",
				fmt.Sprintf("Example: If you leave $%.2f in savings at 5%% APY, you'd earn $%.2f per year.", savingsBalance, savingsBalance*0.05),
				"By withdrawing, you're giving up this passive income.",
				"",
				"✅ I'll allow this withdrawal, BUT...",
				"",
				"To protect your financial health, I'm going to help you set a weekly spending budget.",
				"This will prevent you from needing emergency withdrawals in the future.",
			}...)
		} else {
			// SAFE WITHDRAWAL - Sufficient liquidity for big purchase
			recommendations = append(recommendations, []string{
				"✅ Withdrawal Safety Check:",
				"",
				fmt.Sprintf("Your current situation:"),
				fmt.Sprintf("   • Wallet (liquid): $%.2f (%.0f%% of total)", walletBalance, liquidityRatio*100),
				fmt.Sprintf("   • Savings: $%.2f", savingsBalance),
				fmt.Sprintf("   • Total: $%.2f", totalLiquidity),
				"",
				"✅ This is a safe withdrawal situation:",
				"You have sufficient liquid funds available, so withdrawing from savings is reasonable.",
				"",
				"💡 Quick Financial Education:",
				"",
				"Even though this is safe, here's what you should consider:",
				"   1. Opportunity Cost: Money in savings earns compound interest",
				"   2. Rebuilding: Plan to replenish your savings after this withdrawal",
				"   3. Goals: Make sure this purchase aligns with your financial priorities",
				"",
				"📊 Smart Withdrawal Practices:",
				"   • Only withdraw for planned big purchases or emergencies",
				"   • Try to maintain at least 50% of your wealth in savings/investments",
				"   • Set a goal to replace withdrawn funds within 3 months",
				"",
				fmt.Sprintf("💰 Cost of Withdrawal: At 5%% APY, every $100 withdrawn costs you $5/year in lost earnings."),
				"",
				"✅ You're cleared to proceed with this withdrawal.",
				"Just specify the amount and currency, and I'll help you withdraw.",
			}...)
		}
		
		// Calculate recommended weekly budget for unsafe withdrawals
		var weeklyBudgetLimit float64
		if isUnsafeWithdrawal {
			// Suggest conservative budget: 15% of total liquidity per week
			weeklyBudgetLimit = totalLiquidity * 0.15
			recommendations = append(recommendations, []string{
				"",
				"🎯 Recommended Weekly Budget:",
				fmt.Sprintf("   • Weekly spending limit: $%.2f", weeklyBudgetLimit),
				fmt.Sprintf("   • This keeps you at ~60%% of your income for essentials"),
				"   • Leaves room to rebuild your liquid funds",
				"",
				"Would you like me to set this weekly budget goal for you?",
				"(I can track your spending and alert you when you're approaching the limit)",
			}...)
		}
		
		state.ToolResult = map[string]interface{}{
			"status":                "withdrawal_analyzed",
			"recommendation_type":   "withdrawal_education",
			"recommendations":       recommendations,
			"is_unsafe_withdrawal": isUnsafeWithdrawal,
			"wallet_balance":        walletBalance,
			"savings_balance":       savingsBalance,
			"liquidity_ratio":       liquidityRatio,
			"weekly_budget_limit":   weeklyBudgetLimit,
			"can_withdraw":          true,
		}
		
		state.Conversation["handler_type"] = "withdraw"
		state.Conversation["is_unsafe_withdrawal"] = isUnsafeWithdrawal
		state.Conversation["weekly_budget_limit"] = weeklyBudgetLimit
		
		return nil
	})

	// Deposit - simple deposit to savings with balance check
	graph.AddNode("deposit", func(ctx context.Context, state *GraphState) error {
		log.Println("Deposit: Checking available funds for deposit...")
		
		// Fetch current wallet balance
		balanceRequest := map[string]interface{}{}
		balanceRequestJSON, _ := json.Marshal(balanceRequest)
		
		balanceResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
			UserID:    state.UserID,
			Tool:      "get_balance",
			Input:     balanceRequestJSON,
			RequestID: fmt.Sprintf("deposit-balance-%d", time.Now().Unix()),
		})
		
		if err != nil || !balanceResponse.Success {
			log.Printf("Failed to fetch balance: %v", err)
			state.ToolResult = map[string]interface{}{
				"status":  "error",
				"message": "Unable to check your balance. Please try again.",
			}
			state.Conversation["handler_type"] = "deposit"
			return nil
		}
		
		// Parse wallet balance
		var usdcBalance, eurcBalance float64
		var walletData map[string]interface{}
		if err := json.Unmarshal(balanceResponse.Data, &walletData); err == nil {
			if balances, ok := walletData["balances"].([]interface{}); ok {
				for _, bal := range balances {
					if balMap, ok := bal.(map[string]interface{}); ok {
						if currency, ok := balMap["currency"].(string); ok {
							if balStr, ok := balMap["balance"].(string); ok {
								var amount float64
								fmt.Sscanf(balStr, "%f", &amount)
								if currency == "USD" {
									usdcBalance = amount
								} else if currency == "EUR" {
									eurcBalance = amount
								}
							}
						}
					}
				}
			}
		}
		
		// Fetch vault rates to show APY
		vaultResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
			UserID:    state.UserID,
			Tool:      "get_vault_rates",
			Input:     balanceRequestJSON,
			RequestID: fmt.Sprintf("deposit-rates-%d", time.Now().Unix()),
		})
		
		var usdcAPY, eurcAPY float64
		if err == nil && vaultResponse.Success {
			var vaultData map[string]interface{}
			if err := json.Unmarshal(vaultResponse.Data, &vaultData); err == nil {
				if vaults, ok := vaultData["vaults"].([]interface{}); ok {
					for _, vault := range vaults {
						if vaultMap, ok := vault.(map[string]interface{}); ok {
							if currency, ok := vaultMap["currency"].(string); ok {
								if apyStr, ok := vaultMap["apy"].(string); ok {
									var apy float64
									fmt.Sscanf(apyStr, "%f", &apy)
									if currency == "USD" {
										usdcAPY = apy
									} else if currency == "EUR" {
										eurcAPY = apy
									}
								}
							}
						}
					}
				}
			}
		}
		
		var recommendations []string
		recommendations = append(recommendations, []string{
			"💰 Ready to Deposit into Savings:",
			"",
			"Your Available Balances:",
			fmt.Sprintf("   • USD: $%.2f (earning %.2f%% APY)", usdcBalance, usdcAPY),
			fmt.Sprintf("   • EUR: €%.2f (earning %.2f%% APY)", eurcBalance, eurcAPY),
			"",
			"✨ Benefits of Depositing:",
			"   • Earn passive income through compound interest",
			"   • Your funds are secure and accessible anytime",
			"   • Interest accrues daily and compounds automatically",
			"   • No lock-up periods or penalties for early withdrawal",
			"",
		}...)
		
		// Show potential earnings examples
		if usdcBalance > 0 && usdcAPY > 0 {
			yearlyEarnings := usdcBalance * (usdcAPY / 100)
			monthlyEarnings := yearlyEarnings / 12
			recommendations = append(recommendations, []string{
				"📈 Potential Earnings (USD):",
				fmt.Sprintf("   • If you deposit $%.2f:", usdcBalance),
				fmt.Sprintf("     - Monthly: $%.2f", monthlyEarnings),
				fmt.Sprintf("     - Yearly: $%.2f", yearlyEarnings),
				"",
			}...)
		}
		
		if eurcBalance > 0 && eurcAPY > 0 {
			yearlyEarnings := eurcBalance * (eurcAPY / 100)
			monthlyEarnings := yearlyEarnings / 12
			recommendations = append(recommendations, []string{
				"📈 Potential Earnings (EUR):",
				fmt.Sprintf("   • If you deposit €%.2f:", eurcBalance),
				fmt.Sprintf("     - Monthly: €%.2f", monthlyEarnings),
				fmt.Sprintf("     - Yearly: €%.2f", yearlyEarnings),
				"",
			}...)
		}
		
		recommendations = append(recommendations, []string{
			"💬 How to Deposit:",
			"Just tell me the amount and currency you'd like to deposit.",
			"Examples:",
			"   • \"Deposit 100 USD\"",
			"   • \"Put 50 EUR in savings\"",
			"   • \"Save 200 USD\"",
		}...)
		
		state.ToolResult = map[string]interface{}{
			"status":          "ready_to_deposit",
			"recommendations": recommendations,
			"usd_balance":     usdcBalance,
			"eur_balance":     eurcBalance,
			"usd_apy":         usdcAPY,
			"eur_apy":         eurcAPY,
		}
		
		state.Conversation["handler_type"] = "deposit"
		
		return nil
	})

	// SECOND LAYER NODES - Processing nodes

	// Execute tool based on first layer routing
	graph.AddNode("execute_tool", func(ctx context.Context, state *GraphState) error {
		log.Println("Executing tool based on handler type...")
		handlerType := state.Conversation["handler_type"].(string)
		
		switch handlerType {
		case "image_payment":
			// Process payment splitting
			log.Println("Processing payment splits from receipt...")
		case "financial_help":
			// Fetch and analyze financial data for low funds
			log.Println("Fetching financial data for assistance...")
		case "financial_help_low_funds":
			// Fetch transactions for spending analysis
			log.Println("Fetching transactions for spending analysis...")
		case "spending_analysis":
			// Spending analysis complete, ready for budget recommendations
			log.Println("Spending analysis complete, preparing budget recommendations...")
		case "budget_recommendations":
			// Budget recommendations ready to display
			log.Println("Budget recommendations ready, preparing response...")
		case "withdraw":
			// Withdrawal analysis complete, ready to display education and allow withdrawal
			log.Println("Withdrawal analysis complete, preparing educational response...")
		case "deposit":
			// Deposit info ready, waiting for user to specify amount
			log.Println("Deposit information displayed, ready to process deposit...")
		case "financial_save":
			// Fetch and analyze savings opportunities
			log.Println("Analyzing savings and investment opportunities...")
		case "general":
			// Handle standard queries
			log.Println("Processing general query...")
		}
		
		return nil
	})

	// Generate final response
	graph.AddNode("generate_response", func(ctx context.Context, state *GraphState) error {
		log.Println("Generating final response...")
		handlerType := state.Conversation["handler_type"].(string)
		
		var response string
		switch handlerType {
		case "image_payment":
			result := state.ToolResult.(map[string]interface{})
			response = fmt.Sprintf("I analyzed the receipt. Total: $%.2f. I can split this %d ways ($%.2f each) and send to %v. Ready to proceed?",
				result["total_amount"], result["split_count"], result["per_person"], result["recipients"])
		case "investment_reminder":
			result := state.ToolResult.(map[string]interface{})
			response = "💡 " + result["message_to_user"].(string) + "\n\n"
			response += "I can help you set up reminders for periodic investments:\n\n"
			
			if freqOptions, ok := result["frequency_options"].(map[string]interface{}); ok {
				for freq, details := range freqOptions {
					if detailMap, ok := details.(map[string]interface{}); ok {
						response += fmt.Sprintf("📅 **%s**: %s\n", freq, detailMap["description"])
						response += fmt.Sprintf("   Amount: $%.2f over %s\n\n", detailMap["amount"], detailMap["duration"])
					}
				}
			}
			
			if benefits, ok := result["benefits"].([]string); ok {
				response += "Benefits:\n"
				for _, benefit := range benefits {
					response += fmt.Sprintf("%s\n", benefit)
				}
			}
			
			response += "\nWould you like to set up reminders? If so, which frequency works best for you?"
			
			// Add deposit prompt
			if depositPrompt, ok := result["deposit_now_prompt"].(string); ok {
				response += "\n\n" + depositPrompt
				response += "\n\nJust let me know:\n1️⃣  Your preferred reminder frequency (weekly/bi-weekly/monthly)\n2️⃣  If you want to make a deposit now (and how much)"
			}
		case "financial_save":
			result := state.ToolResult.(map[string]interface{})
			recommendations := result["recommendations"].([]string)
			response = "Great news! You have sufficient funds for saving and investing. Here are my recommendations:\n"
			for _, rec := range recommendations {
				response += fmt.Sprintf("%s\n", rec)
			}
		case "financial_help":
			result := state.ToolResult.(map[string]interface{})
			recommendations := result["recommendations"].([]string)
			response = "I've analyzed your financial situation. Here are some steps to improve your finances:\n"
			for _, rec := range recommendations {
				response += fmt.Sprintf("• %s\n", rec)
			}
		case "spending_analysis":
			result := state.ToolResult.(map[string]interface{})
			recommendations := result["recommendations"].([]string)
			response = ""
			for _, rec := range recommendations {
				response += fmt.Sprintf("%s\n\n", rec)
			}
		case "budget_recommendations":
			result := state.ToolResult.(map[string]interface{})
			recommendations := result["recommendations"].([]string)
			response = ""
			for _, rec := range recommendations {
				response += fmt.Sprintf("%s\n", rec)
			}
		case "withdraw":
			result := state.ToolResult.(map[string]interface{})
			recommendations := result["recommendations"].([]string)
			response = ""
			for _, rec := range recommendations {
				response += fmt.Sprintf("%s\n", rec)
			}
			isUnsafe := result["is_unsafe_withdrawal"].(bool)
			if isUnsafe {
				weeklyLimit := result["weekly_budget_limit"].(float64)
				response += fmt.Sprintf("\n\n💬 Tell me the amount and currency you want to withdraw, and I'll process it.\n")
				response += fmt.Sprintf("After withdrawal, I can help you set up that $%.2f weekly budget to protect your finances.\n", weeklyLimit)
			} else {
				response += "\n\n💬 Just tell me the amount and currency you'd like to withdraw (e.g., \"withdraw 100 USD\").\n"
			}
		case "deposit":
			result := state.ToolResult.(map[string]interface{})
			recommendations := result["recommendations"].([]string)
			response = ""
			for _, rec := range recommendations {
				response += fmt.Sprintf("%s\n", rec)
			}
		case "general":
			response = "I can help you with that. Let me check your account..."
		}
		
		state.Messages = append(state.Messages, response)
		return nil
	})

	// Define the workflow edges
	// Orchestrator routes to one of three first-layer nodes
	graph.AddEdge("orchestrator", "general_inquiry")
	graph.AddEdge("orchestrator", "image_payment")
	graph.AddEdge("orchestrator", "financial_help")
	graph.AddEdge("orchestrator", "withdraw")
	graph.AddEdge("orchestrator", "deposit")
	
	// Financial help branches to save or help route based on balance
	graph.AddEdge("financial_help", "financial_save")
	graph.AddEdge("financial_help", "financial_help_low_funds")
	
	// Financial save leads to investment reminder
	graph.AddEdge("financial_save", "investment_reminder")
	
	// All paths converge to execute_tool
	graph.AddEdge("general_inquiry", "execute_tool")
	graph.AddEdge("image_payment", "execute_tool")
	graph.AddEdge("investment_reminder", "execute_tool")
	graph.AddEdge("withdraw", "execute_tool")
	graph.AddEdge("deposit", "execute_tool")
	graph.AddEdge("financial_help_low_funds", "spending_analysis")
	graph.AddEdge("spending_analysis", "budget_recommendations")
	graph.AddEdge("budget_recommendations", "execute_tool")
	
	// Execute tool leads to response generation
	graph.AddEdge("execute_tool", "generate_response")

	// Set orchestrator as the starting point
	graph.SetStart("orchestrator")

	return graph
}

// RunGraphExample demonstrates how to use the graph workflow
func RunGraphExample(liminalExecutor core.ToolExecutor) {
	graph := CreateFinancialAgentGraph(liminalExecutor)

	ctx := context.Background()
	userID := "example-user-123"

	if err := graph.Execute(ctx, userID); err != nil {
		log.Printf("Graph execution failed: %v", err)
		return
	}

	log.Println("Graph execution completed successfully")
	log.Printf("Final state messages: %v", graph.State.Messages)
}
