import { useEffect, useState } from 'react'

export type ConnectionState = 'connecting' | 'connected' | 'reconnecting'

type PresencePayload = {
  viewers: string[]
}

export function usePresence(viewerName: string, reportId: string) {
  const [viewers, setViewers] = useState<string[]>([])
  const [connectionState, setConnectionState] =
    useState<ConnectionState>('connecting')
  const [activeViewer, setActiveViewer] = useState(viewerName)

  if (activeViewer !== viewerName) {
    setActiveViewer(viewerName)
    setViewers([])
    setConnectionState('connecting')
  }

  useEffect(() => {
    if (!viewerName) {
      return
    }

    const params = new URLSearchParams({ name: viewerName })
    const eventSource = new EventSource(
      `/api/reports/${reportId}/presence?${params}`,
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
  }, [viewerName, reportId])

  return { viewers, connectionState }
}
