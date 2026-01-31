import React from 'react'
import ReactDOM from 'react-dom/client'
import { NimChat } from '@liminalcash/nim-chat'
import '@liminalcash/nim-chat/styles.css'
import './styles.css'
import { SpendingCategories } from './SpendingCategories'

// Weekly Spending Goal Component
function WeeklySpendingGoal({ wsUrl }: { wsUrl: string }) {
  const [goalData, setGoalData] = React.useState<any>(null)
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState<string | null>(null)
  const [rawText, setRawText] = React.useState<string>('')
  const wsRef = React.useRef<WebSocket | null>(null)

  const [authToken, setAuthToken] = React.useState<string | null>(null)

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

  React.useEffect(() => {
    let reconnectTimeout: NodeJS.Timeout | null = null
    let responseTimeout: NodeJS.Timeout | null = null

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
          console.log('WebSocket closed')
          // Do not auto-reconnect in a loop; refresh manually if needed
        }
      } catch (e) {
        console.error('WebSocket connection failed:', e)
        setError('Connection failed')
        setLoading(false)
      }
    }

    connectAndFetch()

    return () => {
      if (reconnectTimeout) clearTimeout(reconnectTimeout)
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

  return (
    <>
      <main>
        <h1>Build financial autonomy for AI</h1>

        <WeeklySpendingGoal wsUrl={wsUrl} />
        <SpendingCategories wsUrl={wsUrl} />

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
    </>
  )
}

ReactDOM.createRoot(document.getElementById('root')!).render(<App />)
