import React from 'react'

interface ChartData {
  labels: string[]
  values: number[]
  title: string
  chart_type?: string
}

interface NimOutputProps {
  wsUrl: string
  chartData: any
}

export function NimOutput({ wsUrl, chartData: chartDataFromProps }: NimOutputProps) {
  const [chartData, setChartData] = React.useState<ChartData | null>(null)
  const [loading, setLoading] = React.useState(false)
  const wsRef = React.useRef<WebSocket | null>(null)

  // Update local state when props change
  React.useEffect(() => {
    if (chartDataFromProps) {
      console.log('📊 Chart data received from props:', chartDataFromProps)
      setChartData(chartDataFromProps)
    }
  }, [chartDataFromProps])

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

  const requestChart = React.useCallback(() => {
    console.log('📊 Manual refresh - requesting chart...')
    setLoading(true)
    
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      console.log('📊 Opening new WebSocket connection...')
      const url = buildWsUrl()
      wsRef.current = new WebSocket(url)
      
      wsRef.current.onopen = () => {
        console.log('📊 WebSocket opened, sending chart request')
        wsRef.current?.send(JSON.stringify({ type: 'new_conversation' }))
        setTimeout(() => {
          const message = {
            type: 'message',
            content: 'Show me a chart'
          }
          wsRef.current?.send(JSON.stringify(message))
        }, 500)
      }

      wsRef.current.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          console.log('📊 Received message:', data)
          
          if (data.type === 'text' && typeof data.content === 'string') {
            console.log('📊 Text content:', data.content)
            const chartMatch = data.content.match(/CHART_DATA_START\s*(\{[\s\S]*?\})\s*CHART_DATA_END/i)
            if (chartMatch) {
              console.log('📊 Found chart markers!')
              console.log('📊 Matched content:', chartMatch[1])
              try {
                const chartInfo = JSON.parse(chartMatch[1])
                console.log('📊 Parsed chart info:', chartInfo)
                if (chartInfo.labels && chartInfo.values && chartInfo.title) {
                  console.log('📊 Setting chart data')
                  setChartData(chartInfo)
                  setLoading(false)
                } else {
                  console.error('📊 Chart info missing required fields:', chartInfo)
                  setLoading(false)
                }
              } catch (e) {
                console.error('Failed to parse chart data:', e)
                setLoading(false)
              }
            } else {
              console.log('📊 No CHART_DATA_START markers found in text')
            }
          }
          
          // Also check for tool results
          if (data.type === 'tool_result') {
            console.log('📊 Tool result received:', data.tool_name, data)
            if (data.tool_name === 'generate_chart' && data.result && data.result.chart_data) {
              console.log('📊 Chart from generate_chart tool:', data.result.chart_data)
              setChartData(data.result.chart_data)
              setLoading(false)
            }
          }
        } catch (e) {
          console.error('Failed to parse message:', e)
        }
      }

      wsRef.current.onerror = (error) => {
        console.error('📊 WebSocket error:', error)
        setLoading(false)
      }

      wsRef.current.onclose = () => {
        console.log('📊 WebSocket closed')
        setLoading(false)
      }
    } else {
      console.log('📊 Using existing WebSocket connection')
      const message = {
        type: 'message',
        content: 'Show me a chart'
      }
      wsRef.current.send(JSON.stringify(message))
    }
  }, [wsUrl, buildWsUrl])

  React.useEffect(() => {
    return () => {
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [])

  if (!chartData) {
    return (
      <div className="nim-output-widget">
        <h3>📊 Balance Trend</h3>
        <p className="hint">Your account balance trend will appear here</p>
        <p className="hint-small">Try asking: "Show me my balance trend"</p>
        <button 
          className="refresh-chart-btn"
          onClick={requestChart}
          disabled={loading}
        >
          {loading ? 'Loading...' : '🔄 Test Chart'}
        </button>
      </div>
    )
  }

  // Render line chart showing balance over time
  const padding = 40
  const width = 500
  const height = 300
  const chartWidth = width - padding * 2
  const chartHeight = height - padding * 2
  
  const maxValue = Math.max(...chartData.values)
  const minValue = Math.min(...chartData.values)
  const valueRange = maxValue - minValue || 1 // Avoid division by zero
  
  const points = chartData.values.map((value, index) => {
    const x = padding + (index / Math.max(chartData.values.length - 1, 1)) * chartWidth
    const y = padding + chartHeight - ((value - minValue) / valueRange) * chartHeight
    return `${x},${y}`
  }).join(' ')
  
  return (
    <div className="nim-output-widget">
      <h3>📊 {chartData.title}</h3>
      
      <div className="line-chart-container">
        <svg viewBox={`0 0 ${width} ${height}`} className="line-chart">
          {/* Grid lines */}
          {[0, 0.25, 0.5, 0.75, 1].map((factor, i) => {
            const yValue = minValue + (maxValue - minValue) * (1 - factor)
            return (
              <g key={i}>
                <line
                  x1={padding}
                  y1={padding + chartHeight * factor}
                  x2={width - padding}
                  y2={padding + chartHeight * factor}
                  stroke="#E0E0E0"
                  strokeWidth="1"
                />
                <text
                  x={padding - 5}
                  y={padding + chartHeight * factor + 4}
                  textAnchor="end"
                  fontSize="10"
                  fill="#666"
                >
                  ${yValue.toFixed(0)}
                </text>
              </g>
            )
          })}
          
          {/* Line */}
          <polyline
            points={points}
            fill="none"
            stroke="#4ECDC4"
            strokeWidth="3"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
          
          {/* Points and labels */}
          {chartData.values.map((value, index) => {
            const x = padding + (index / Math.max(chartData.values.length - 1, 1)) * chartWidth
            const y = padding + chartHeight - ((value - minValue) / valueRange) * chartHeight
            
            // Only show labels for every Nth point if there are many points
            const showLabel = chartData.values.length <= 10 || index % Math.ceil(chartData.values.length / 10) === 0
            
            return (
              <g key={index}>
                <circle
                  cx={x}
                  cy={y}
                  r="4"
                  fill="#4ECDC4"
                  stroke="white"
                  strokeWidth="2"
                />
                {showLabel && (
                  <>
                    <text
                      x={x}
                      y={height - 5}
                      textAnchor="middle"
                      fontSize="9"
                      fill="#666"
                      transform={`rotate(-45, ${x}, ${height - 5})`}
                    >
                      {chartData.labels[index]}
                    </text>
                    <title>${value.toFixed(2)}</title>
                  </>
                )}
              </g>
            )
          })}
        </svg>
      </div>
      
      <button 
        className="clear-chart-btn"
        onClick={() => setChartData(null)}
      >
        Clear
      </button>
    </div>
  )
}
