import React from 'react'
import ReactDOM from 'react-dom/client'
import { NimChat } from '@liminalcash/nim-chat'
import '@liminalcash/nim-chat/styles.css'
import './styles.css'
import { SpendingCategories } from './SpendingCategories'

// Shared context for weekly goal data
const WeeklyGoalContext = React.createContext<any>(null)

// Notification Bell Component - reads from shared context
function NotificationBell() {
  const goalData = React.useContext(WeeklyGoalContext)
  const [showNotification, setShowNotification] = React.useState(false)
  const [hasAutoShown, setHasAutoShown] = React.useState(false)

  // Auto-show notification when budget warning or exceeded
  React.useEffect(() => {
    if (!goalData || !goalData.goal_set || hasAutoShown) return
    
    const percentage = goalData.percentage || 0
    if (percentage >= 80) {
      setShowNotification(true)
      setHasAutoShown(true)
      
      // Auto-hide after 8 seconds
      const timer = setTimeout(() => setShowNotification(false), 8000)
      return () => clearTimeout(timer)
    }
  }, [goalData, hasAutoShown])

  if (!goalData || !goalData.goal_set) {
    return (
      <>
        <div className="notification-bell" onClick={() => setShowNotification(!showNotification)}>
          <span className="bell-icon">🔔</span>
        </div>
        {showNotification && (
          <div className="notification-popup">
            <div className="notification-header normal">
              <span>💰 Weekly Budget</span>
              <button onClick={() => setShowNotification(false)}>×</button>
            </div>
            <div className="notification-body">
              <p>No weekly spending goal set yet</p>
              <p className="notification-details">
                Ask Nim: "Set a weekly spend of 5 LIL"
              </p>
            </div>
          </div>
        )}
      </>
    )
  }

  const percentage = goalData.percentage || 0
  const budgetExceeded = percentage >= 100
  const budgetWarning = percentage >= 80 && percentage < 100

  return (
    <>
      <div className="notification-bell" onClick={() => setShowNotification(!showNotification)}>
        <span className="bell-icon">🔔</span>
        {(budgetExceeded || budgetWarning) && <span className="notification-badge"></span>}
      </div>
      
      {showNotification && (
        <div className="notification-popup">
          <div className={`notification-header ${budgetExceeded ? 'exceeded' : budgetWarning ? 'warning' : 'normal'}`}>
            <span>
              {budgetExceeded ? '🚨 Budget Exceeded!' : budgetWarning ? '⚠️ Budget Warning' : '💰 Weekly Budget'}
            </span>
            <button onClick={() => setShowNotification(false)}>×</button>
          </div>
          <div className="notification-body">
            {budgetExceeded ? (
              <>
                <p><strong>You've exceeded your weekly spending limit!</strong></p>
                <p className="notification-details">
                  Spent: {goalData.spent_so_far} {goalData.currency}<br/>
                  Goal: {goalData.goal_amount} {goalData.currency}<br/>
                  Over by: {(goalData.spent_so_far - goalData.goal_amount).toFixed(2)} {goalData.currency}
                </p>
              </>
            ) : budgetWarning ? (
              <>
                <p><strong>You're approaching your spending limit!</strong></p>
                <p className="notification-details">
                  Spent: {goalData.spent_so_far} {goalData.currency}<br/>
                  Goal: {goalData.goal_amount} {goalData.currency}<br/>
                  Remaining: {goalData.remaining} {goalData.currency}
                </p>
              </>
            ) : (
              <>
                <p>Your weekly budget is on track ✓</p>
                <p className="notification-details">
                  Spent: {goalData.spent_so_far} {goalData.currency}<br/>
                  Goal: {goalData.goal_amount} {goalData.currency}<br/>
                  Remaining: {goalData.remaining} {goalData.currency}
                </p>
              </>
            )}
          </div>
        </div>
      )}
    </>
  )
}

// Weekly Spending Goal Component
function WeeklySpendingGoalWithContext({ wsUrl, setGoalData: setParentGoalData, onRefresh }: { wsUrl: string, setGoalData: (data: any) => void, onRefresh?: () => React.MutableRefObject<(() => void) | null> }) {
  const [goalData, setGoalData] = React.useState<any>(null)
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState<string | null>(null)
  const [rawText, setRawText] = React.useState<string>('')
  const wsRef = React.useRef<WebSocket | null>(null)

  const [authToken, setAuthToken] = React.useState<string | null>(null)

  // Update parent context whenever goal data changes
  React.useEffect(() => {
    setParentGoalData(goalData)
  }, [goalData, setParentGoalData])

  const buildWsUrl = React.useCallback(() => {
    try {
      const token = localStorage.getItem('nim_access_token')
      setAuthToken(token)
      if (!token) return wsUrl
      const url = new URL(wsUrl)
      url.searchParams.set('token', token)
      return url.toString()
    } catch {
      return wsUrl
    }
  }, [wsUrl])

  const requestProgress = React.useCallback(() => {
    const message = {
      type: 'message',
      content: 'Call get_weekly_spending_progress and respond with WEEKLY_SPEND_PROGRESS JSON only.'
    }
    wsRef.current?.send(JSON.stringify(message))
  }, [])

  const parseProgressFromText = React.useCallback((text: string) => {
    // Check for WEEKLY_SPEND_PROGRESS marker
    const markerMatch = text.match(/WEEKLY_SPEND_PROGRESS\s*:\s*(\{.*\})/)
    if (markerMatch) {
      try {
        return JSON.parse(markerMatch[1])
      } catch {
        return null
      }
    }

    // Strip markdown code blocks (```json ... ```)
    let cleaned = text.trim()
    if (cleaned.startsWith('```')) {
      // Remove opening ```json or ```
      cleaned = cleaned.replace(/^```(?:json)?\s*/i, '')
      // Remove closing ```
      cleaned = cleaned.replace(/\s*```$/, '')
      cleaned = cleaned.trim()
    }

    // Try to parse as JSON
    if (cleaned.startsWith('{') && cleaned.endsWith('}')) {
      try {
        return JSON.parse(cleaned)
      } catch {
        return null
      }
    }

    return null
  }, [])

  const fetchGoalData = React.useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      // Re-initialize conversation to get fresh data
      wsRef.current.send(JSON.stringify({ type: 'new_conversation' }))
    }
  }, [])

  React.useEffect(() => {
    if (onRefresh) {
      const ref = onRefresh()
      ref.current = fetchGoalData
    }
  }, [onRefresh, fetchGoalData])

  React.useEffect(() => {
    if (onRefresh) {
      // Expose refresh function to parent
      onRefresh.call = fetchGoalData
    }
  }, [onRefresh, fetchGoalData])

  React.useEffect(() => {
    let responseTimeout: NodeJS.Timeout | null = null
    let pollInterval: NodeJS.Timeout | null = null

    const connectAndFetch = () => {
      try {
        const url = buildWsUrl()
        wsRef.current = new WebSocket(url)

        wsRef.current.onopen = () => {
          console.log('WeeklyGoal WS open:', url)
          wsRef.current?.send(JSON.stringify({ type: 'new_conversation' }))

          responseTimeout = setTimeout(() => {
            if (loading) {
              setLoading(false)
              setError('No goal set')
            }
          }, 5000)

          // Poll for updates every 30 seconds
          pollInterval = setInterval(() => {
            if (wsRef.current?.readyState === WebSocket.OPEN) {
              console.log('Polling for weekly goal updates...')
              wsRef.current.send(JSON.stringify({ type: 'new_conversation' }))
            }
          }, 30000)
        }

        wsRef.current.onmessage = (event) => {
          try {
            console.log('WeeklyGoal WS raw:', event.data)
            const data = JSON.parse(event.data)
            if (data.type === 'conversation_started') {
              requestProgress()
              return
            }
            if (data.type === 'text' && typeof data.content === 'string') {
              setRawText(data.content)
              const parsed = parseProgressFromText(data.content)
              if (parsed) {
                console.log('Weekly goal data (parsed):', parsed)
                // Clear the timeout since we got data
                if (responseTimeout) clearTimeout(responseTimeout)
                setGoalData(parsed)
                setLoading(false)
                setError(null)
              } else if (data.content.toLowerCase().includes('no weekly spending goal')) {
                console.log('Weekly goal data: no goal set')
                if (responseTimeout) clearTimeout(responseTimeout)
                setGoalData({ goal_set: false })
                setLoading(false)
                setError(null)
              }
            } else if (data.type === 'error') {
              console.error('WeeklyGoal WS error:', data)
              if (responseTimeout) clearTimeout(responseTimeout)
              setError(data.content || 'Failed to fetch goal')
              setLoading(false)
            }
          } catch (e) {
            console.error('Failed to parse message:', e)
          }
        }

        wsRef.current.onerror = (err) => {
          console.error('WebSocket error:', err)
          setError('Connection error')
          setLoading(false)
        }

        wsRef.current.onclose = () => {
          console.log('WebSocket closed, reconnecting...')
          // Reconnect after 2 seconds
          setTimeout(() => {
            connectAndFetch()
          }, 2000)
        }
      } catch (e) {
        console.error('WebSocket connection failed:', e)
        setError('Connection failed')
        setLoading(false)
      }
    }

    connectAndFetch()

    return () => {
      if (pollInterval) clearInterval(pollInterval)
      if (responseTimeout) clearTimeout(responseTimeout)
      wsRef.current?.close()
    }
  }, [buildWsUrl, requestProgress, parseProgressFromText])

  if (loading) {
    return (
      <div className="weekly-goal-widget loading">
        <div className="spinner"></div>
      </div>
    )
  }

  if (!authToken) {
    return (
      <div className="weekly-goal-widget no-goal">
        <h3>💰 Weekly Spending Goal</h3>
        <p>Login required to load weekly goal.</p>
        <p className="hint">Open Nim chat and log in first.</p>
      </div>
    )
  }

  if (error || !goalData) {
    return (
      <div className="weekly-goal-widget no-goal">
        <h3>💰 Weekly Spending Goal</h3>
        <p>No goal set yet</p>
        <p className="hint">Ask Nim: "Set a weekly spend of 5 LIL"</p>
      </div>
    )
  }

  if (!goalData.goal_set) {
    return (
      <div className="weekly-goal-widget no-goal">
        <h3>💰 Weekly Spending Goal</h3>
        <p>No goal set yet</p>
        <p className="hint">Ask Nim: "Set a weekly spend of 5 LIL"</p>
      </div>
    )
  }

  const { spent_so_far, goal_amount, remaining, percentage, on_track, days_left, currency } = goalData
  const progressWidth = Math.min(percentage, 100)
  const statusClass = percentage >= 100 ? 'over-budget' : percentage >= 80 ? 'warning' : 'on-track'

  return (
    <div className="weekly-goal-widget">
      <h3>💰 Weekly Spending Goal</h3>
      
      <div className="spending-details">
        <div className="spent">
          <span className="label">Spent This Week</span>
          <span className="value">{spent_so_far} {currency}</span>
        </div>
        <div className="remaining">
          <span className="label">Remaining</span>
          <span className={`value ${remaining < 0 ? 'over' : ''}`}>
            {remaining} {currency}
          </span>
        </div>
      </div>

      <div className="progress-container">
        <div className="progress-bar">
          <div 
            className={`progress-fill ${statusClass}`}
            style={{ width: `${progressWidth}%` }}
          >
            <span className="progress-text">{Number(percentage).toFixed(2)}%</span>
          </div>
        </div>
      </div>

      <div className={`status-badge ${on_track ? 'on-track' : 'off-track'}`}>
        {on_track 
          ? `✓ On track · ${days_left} day${days_left !== 1 ? 's' : ''} left`
          : `⚠ Over budget · ${days_left} day${days_left !== 1 ? 's' : ''} left`
        }
      </div>
    </div>
  )
}

function App() {
  const wsUrl = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws'
  const apiUrl = import.meta.env.VITE_API_URL || 'https://api.liminal.cash'
  const [weeklyGoalData, setWeeklyGoalData] = React.useState<any>(null)
  const weeklyGoalRefreshRef = React.useRef<(() => void) | null>(null)
  const categoriesRefreshRef = React.useRef<(() => void) | null>(null)
  const lastTransactionCountRef = React.useRef<number | null>(null)
  const monitorWsRef = React.useRef<WebSocket | null>(null)

  // Monitor for new transactions and trigger widget refreshes
  React.useEffect(() => {
    const buildWsUrl = () => {
      try {
        const token = localStorage.getItem('nim_access_token')
        if (!token) return wsUrl
        const url = new URL(wsUrl)
        url.searchParams.set('token', token)
        return url.toString()
      } catch {
        return wsUrl
      }
    }

    const connectMonitor = () => {
      const url = buildWsUrl()
      monitorWsRef.current = new WebSocket(url)

      monitorWsRef.current.onopen = () => {
        console.log('Transaction monitor connected')
        // Initialize conversation and check transactions
        monitorWsRef.current?.send(JSON.stringify({ type: 'new_conversation' }))
        
        // Poll every 10 seconds for new transactions
        const pollInterval = setInterval(() => {
          if (monitorWsRef.current?.readyState === WebSocket.OPEN) {
            const message = {
              type: 'message',
              content: 'Call get_transactions and return just the count of total transactions as a number'
            }
            monitorWsRef.current.send(JSON.stringify(message))
          }
        }, 10000)

        monitorWsRef.current.addEventListener('close', () => {
          clearInterval(pollInterval)
        })
      }

      monitorWsRef.current.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          
          if (data.type === 'conversation_started') {
            // Request initial transaction count
            const message = {
              type: 'message',
              content: 'Call get_transactions and return just the count of total transactions as a number'
            }
            monitorWsRef.current?.send(JSON.stringify(message))
            return
          }
          
          if (data.type === 'text' && typeof data.content === 'string') {
            // Try to extract transaction count from response
            const countMatch = data.content.match(/(?:count|total|transactions?)[:\s]*([0-9]+)/i)
            if (countMatch) {
              const currentCount = parseInt(countMatch[1], 10)
              
              if (lastTransactionCountRef.current !== null && currentCount > lastTransactionCountRef.current) {
                console.log('🔔 New transaction detected! Refreshing widgets...')
                // Trigger refresh of all widgets
                weeklyGoalRefreshRef.current?.()
                categoriesRefreshRef.current?.()
              }
              
              lastTransactionCountRef.current = currentCount
            }
          }
        } catch (e) {
          console.error('Monitor message parse error:', e)
        }
      }

      monitorWsRef.current.onclose = () => {
        console.log('Transaction monitor disconnected, reconnecting...')
        setTimeout(connectMonitor, 3000)
      }

      monitorWsRef.current.onerror = (err) => {
        console.error('Monitor WebSocket error:', err)
      }
    }

    connectMonitor()

    return () => {
      monitorWsRef.current?.close()
    }
  }, [wsUrl])

  return (
    <WeeklyGoalContext.Provider value={weeklyGoalData}>
      <NotificationBell />
      
      <main>
        <h1>Build financial autonomy for AI</h1>

        <WeeklySpendingGoalWithContext 
          wsUrl={wsUrl} 
          setGoalData={setWeeklyGoalData}
          onRefresh={() => weeklyGoalRefreshRef}
        />
        <SpendingCategories 
          wsUrl={wsUrl}
          onRefresh={() => categoriesRefreshRef}
        />

        <ol>
          <li>
            Download <a href="https://apps.apple.com/app/testflight/id899247664" target="_blank" rel="noopener noreferrer">TestFlight</a> from the App Store
          </li>

          <li>
            Install <a href="https://testflight.apple.com/join/ZYTDH2bd" target="_blank" rel="noopener noreferrer">Liminal via TestFlight</a>
          </li>

          <li>
            Sign up to Liminal (this is how you authenticate with Nim)
          </li>

          <li>
            Clone the <a href="https://github.com/becomeliminal/nim-go-sdk" target="_blank" rel="noopener noreferrer">Nim Go SDK</a>
            <div className="code-block">
              git clone https://github.com/becomeliminal/nim-go-sdk.git<br />
              cd nim-go-sdk/examples/hackathon-starter
            </div>
          </li>

          <li>
            Create a frontend using the <a href="https://github.com/becomeliminal/nim-chat" target="_blank" rel="noopener noreferrer">Nim Chat</a> component (or use this one)
            <div className="code-block">
              cd frontend<br />
              npm install<br />
              npm run dev
            </div>
          </li>

          <li>
            Create a backend using the Nim Go SDK — see the <a href="https://github.com/becomeliminal/nim-go-sdk/tree/master/examples/hackathon-starter" target="_blank" rel="noopener noreferrer">example</a>
            <div className="code-block">
              {`# In a new terminal`}<br />
              cd ..<br />
              cp .env.example .env<br />
              {`# Add your ANTHROPIC_API_KEY to .env`}<br />
              go run main.go
            </div>
          </li>

          <li>
            Build cool tools for Nim
          </li>
        </ol>
      </main>

      <NimChat
        wsUrl={wsUrl}
        apiUrl={apiUrl}
        title="Nim"
        position="bottom-right"
        defaultOpen={false}
      />
    </WeeklyGoalContext.Provider>
  )
}

ReactDOM.createRoot(document.getElementById('root')!).render(<App />)
