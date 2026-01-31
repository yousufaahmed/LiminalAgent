import React from 'react'

interface CategoryData {
  food: number
  travel: number
  subscription: number
  entertainment: number
  electronics: number
  miscellaneous: number
}

interface SpendingCategoriesProps {
  wsUrl: string
}

export function SpendingCategories({ wsUrl }: SpendingCategoriesProps) {
  const [categoryData, setCategoryData] = React.useState<CategoryData | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState<string | null>(null)
  const wsRef = React.useRef<WebSocket | null>(null)

  const buildWsUrl = React.useCallback(() => {
    try {
      const token = localStorage.getItem('nim_access_token')
      if (!token) return wsUrl
      const url = new URL(wsUrl)
      url.searchParams.set('token', token)
      return url.toString()
    } catch {
      return wsUrl
    }
  }, [wsUrl])

  const requestCategories = React.useCallback(() => {
    const message = {
      type: 'message',
      content: 'Call categorize_transactions and return the categories JSON'
    }
    wsRef.current?.send(JSON.stringify(message))
  }, [])

  const parseCategoryData = React.useCallback((text: string) => {
    console.log('Parsing category text:', text)
    
    // Extract JSON from markdown code blocks
    const jsonMatch = text.match(/```(?:json)?\s*(\{[\s\S]*?\})\s*```/)
    if (jsonMatch) {
      try {
        const parsed = JSON.parse(jsonMatch[1])
        console.log('Parsed from markdown:', parsed)
        if (parsed.categories) {
          return parsed.categories
        }
        if ('food' in parsed || 'travel' in parsed) {
          return parsed
        }
      } catch (e) {
        console.error('Failed to parse markdown JSON:', e)
      }
    }

    // Try to find JSON object in text
    const jsonObjectMatch = text.match(/\{[\s\S]*"categories"[\s\S]*?\}(?=\s*\n|$)/m)
    if (jsonObjectMatch) {
      try {
        const parsed = JSON.parse(jsonObjectMatch[0])
        console.log('Parsed from text search:', parsed)
        if (parsed.categories) {
          return parsed.categories
        }
      } catch (e) {
        console.error('Failed to parse extracted JSON:', e)
      }
    }

    // Last resort: try parsing the whole text
    try {
      const parsed = JSON.parse(text.trim())
      console.log('Parsed as whole text:', parsed)
      if (parsed.categories) {
        return parsed.categories
      }
      if ('food' in parsed || 'travel' in parsed) {
        return parsed
      }
    } catch {}

    return null
  }, [])

  React.useEffect(() => {
    let responseTimeout: NodeJS.Timeout | null = null

    const connectAndFetch = () => {
      try {
        const url = buildWsUrl()
        wsRef.current = new WebSocket(url)

        wsRef.current.onopen = () => {
          console.log('Categories WS open:', url)
          wsRef.current?.send(JSON.stringify({ type: 'new_conversation' }))

          responseTimeout = setTimeout(() => {
            if (loading) {
              setLoading(false)
              setError('Failed to load categories')
            }
          }, 10000)
        }

        wsRef.current.onmessage = (event) => {
          try {
            console.log('Categories WS raw:', event.data)
            const data = JSON.parse(event.data)
            
            if (data.type === 'conversation_started') {
              requestCategories()
              return
            }
            
            if (data.type === 'text' && typeof data.content === 'string') {
              const parsed = parseCategoryData(data.content)
              if (parsed) {
                console.log('Category data (parsed):', parsed)
                if (responseTimeout) clearTimeout(responseTimeout)
                setCategoryData(parsed)
                setLoading(false)
                setError(null)
              }
            } else if (data.type === 'error') {
              console.error('Categories WS error:', data)
              if (responseTimeout) clearTimeout(responseTimeout)
              setError(data.content || 'Failed to fetch categories')
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
        }
      } catch (e) {
        console.error('WebSocket connection failed:', e)
        setError('Connection failed')
        setLoading(false)
      }
    }

    connectAndFetch()

    return () => {
      if (responseTimeout) clearTimeout(responseTimeout)
      wsRef.current?.close()
    }
  }, [buildWsUrl, requestCategories, parseCategoryData])

  if (loading) {
    return (
      <div className="spending-categories-widget loading">
        <h3>📊 Spending Categories</h3>
        <div className="spinner"></div>
      </div>
    )
  }

  if (error || !categoryData) {
    return (
      <div className="spending-categories-widget no-data">
        <h3>📊 Spending Categories</h3>
        <p>No spending data available</p>
        <p className="hint">Make some transactions to see category breakdown</p>
      </div>
    )
  }

  // Calculate total and percentages
  const total = Object.values(categoryData).reduce((sum, val) => sum + val, 0)
  
  if (total === 0) {
    return (
      <div className="spending-categories-widget no-data">
        <h3>📊 Spending Categories</h3>
        <p>No spending tracked yet</p>
        <p className="hint">Your spending will be categorized here</p>
      </div>
    )
  }

  const categoryColors: Record<keyof CategoryData, string> = {
    food: '#FF6B6B',
    travel: '#4ECDC4',
    subscription: '#45B7D1',
    entertainment: '#FFA07A',
    electronics: '#98D8C8',
    miscellaneous: '#C7CEEA'
  }

  const categoryIcons: Record<keyof CategoryData, string> = {
    food: '🍔',
    travel: '✈️',
    subscription: '📱',
    entertainment: '🎬',
    electronics: '💻',
    miscellaneous: '📦'
  }

  return (
    <div className="spending-categories-widget">
      <h3>📊 Spending Categories</h3>
      
      <div className="bubble-chart">
        {(Object.keys(categoryData) as Array<keyof CategoryData>).map((category) => {
          const count = categoryData[category]
          if (count === 0) return null
          
          const percentage = (count / total) * 100
          const size = Math.max(60, Math.min(160, 60 + (percentage * 2))) // Size between 60-160px
          
          return (
            <div
              key={category}
              className="bubble"
              style={{
                width: `${size}px`,
                height: `${size}px`,
                backgroundColor: categoryColors[category],
              }}
              title={`${category}: ${count} transactions (${percentage.toFixed(1)}%)`}
            >
              <div className="bubble-content">
                <span className="bubble-icon">{categoryIcons[category]}</span>
                <span className="bubble-label">{category}</span>
                <span className="bubble-count">{count}</span>
              </div>
            </div>
          )
        })}
      </div>

      <div className="category-legend">
        {(Object.keys(categoryData) as Array<keyof CategoryData>).map((category) => {
          const count = categoryData[category]
          if (count === 0) return null
          const percentage = (count / total) * 100
          
          return (
            <div key={category} className="legend-item">
              <span
                className="legend-color"
                style={{ backgroundColor: categoryColors[category] }}
              ></span>
              <span className="legend-text">
                {categoryIcons[category]} {category}: {count} ({percentage.toFixed(1)}%)
              </span>
            </div>
          )
        })}
      </div>
    </div>
  )
}
