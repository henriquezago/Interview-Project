import type { ConnectionState } from '../hooks/usePresence'

type ReportPresenceProps = {
  title: string
  viewerName: string
  viewers: string[]
  connectionState: ConnectionState
  onLeave: () => void
}

export function ReportPresence({
  title,
  viewerName,
  viewers,
  connectionState,
  onLeave,
}: ReportPresenceProps) {
  return (
    <section className="card report" aria-labelledby="report-title">
      <header>
        <div>
          <p className="eyebrow">Report</p>
          <h1 id="report-title">{title}</h1>
        </div>
        <button className="secondary" type="button" onClick={onLeave}>
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
  )
}
