import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'

const report = {
  id: 'weekly-performance',
  title: 'Weekly Performance Report',
}

type ConnectionState = 'connecting' | 'connected' | 'reconnecting'

type PresencePayload = {
  viewers: string[]
}

function App() {
  const [nameInput, setNameInput] = useState('')
  const [viewerName, setViewerName] = useState('')
  const [viewers, setViewers] = useState<string[]>([])
  const [connectionState, setConnectionState] =
    useState<ConnectionState>('connecting')

  useEffect(() => {
    if (!viewerName) return

    const params = new URLSearchParams({ name: viewerName })
    const eventSource = new EventSource(
      `/api/reports/${report.id}/presence?${params}`,
    )

    eventSource.onopen = () => setConnectionState('connected')
    eventSource.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data) as PresencePayload
        if (Array.isArray(payload.viewers)) {
          setViewers(payload.viewers)
        }
      } catch {
        setConnectionState('reconnecting')
      }
    }
    eventSource.onerror = () => setConnectionState('reconnecting')

    return () => eventSource.close()
  }, [viewerName])

  const joinReport = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const name = nameInput.trim()
    if (name) {
      setConnectionState('connecting')
      setViewerName(name)
    }
  }

  const leaveReport = () => {
    setViewerName('')
    setViewers([])
  }

  return (
    <main>
      {!viewerName ? (
        <section className="card sign-in" aria-labelledby="welcome-title">
          <p className="eyebrow">HYPE10 Reports</p>
          <h1 id="welcome-title">Open the report</h1>
          <p className="description">
            Enter your name so teammates can see that you are viewing it.
          </p>
          <form onSubmit={joinReport}>
            <label htmlFor="name">Your name</label>
            <input
              id="name"
              maxLength={50}
              value={nameInput}
              onChange={(event) => setNameInput(event.target.value)}
              placeholder="e.g. Cole"
              autoComplete="name"
              autoFocus
              required
            />
            <button type="submit">View report</button>
          </form>
        </section>
      ) : (
        <section className="card report" aria-labelledby="report-title">
          <header>
            <div>
              <p className="eyebrow">Report</p>
              <h1 id="report-title">{report.title}</h1>
            </div>
            <button className="secondary" type="button" onClick={leaveReport}>
              Leave
            </button>
          </header>

          <div className="presence-heading">
            <h2>Currently viewing</h2>
            <span
              className={`status ${connectionState}`}
              role="status"
              aria-live="polite"
            >
              {connectionState === 'connected' ? 'Live' : connectionState}
            </span>
          </div>

          {viewers.length > 0 ? (
            <ul className="viewer-list">
              {viewers.map((name) => (
                <li key={name}>
                  <span className="avatar" aria-hidden="true">
                    {name.charAt(0).toUpperCase()}
                  </span>
                  <span>{name}</span>
                  {name === viewerName && <small>You</small>}
                </li>
              ))}
            </ul>
          ) : (
            <p className="empty-state">Waiting for viewer information…</p>
          )}
        </section>
      )}
    </main>
  )
}

export default App
