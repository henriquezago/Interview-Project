import { act, renderHook } from '@testing-library/react'
import { usePresence } from './usePresence'

class MockEventSource {
  static instances: MockEventSource[] = []

  url: string
  onopen: ((event: Event) => void) | null = null
  onmessage: ((event: MessageEvent) => void) | null = null
  onerror: ((event: Event) => void) | null = null
  close = vi.fn()

  constructor(url: string) {
    this.url = url
    MockEventSource.instances.push(this)
  }
}

beforeEach(() => {
  MockEventSource.instances = []
  vi.stubGlobal('EventSource', MockEventSource)
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('usePresence', () => {
  it('does not open EventSource when viewerName is empty', () => {
    const { result } = renderHook(() => usePresence('', 'weekly-performance'))

    expect(MockEventSource.instances).toHaveLength(0)
    expect(result.current.viewers).toEqual([])
  })

  it('opens EventSource for the report and starts connecting', () => {
    const { result } = renderHook(() =>
      usePresence('Cole', 'weekly-performance'),
    )

    expect(MockEventSource.instances).toHaveLength(1)
    expect(MockEventSource.instances[0].url).toBe(
      '/api/reports/weekly-performance/presence?name=Cole',
    )
    expect(result.current.connectionState).toBe('connecting')
  })

  it('sets connected on open', () => {
    const { result } = renderHook(() =>
      usePresence('Cole', 'weekly-performance'),
    )

    act(() => {
      MockEventSource.instances[0].onopen?.(new Event('open'))
    })

    expect(result.current.connectionState).toBe('connected')
  })

  it('updates viewers from a valid message', () => {
    const { result } = renderHook(() =>
      usePresence('Cole', 'weekly-performance'),
    )

    act(() => {
      MockEventSource.instances[0].onmessage?.({
        data: JSON.stringify({ viewers: ['Cole', 'Ada'] }),
      } as MessageEvent)
    })

    expect(result.current.viewers).toEqual(['Cole', 'Ada'])
  })

  it('sets reconnecting when a message is invalid JSON', () => {
    const { result } = renderHook(() =>
      usePresence('Cole', 'weekly-performance'),
    )

    act(() => {
      MockEventSource.instances[0].onmessage?.({
        data: 'not-json',
      } as MessageEvent)
    })

    expect(result.current.connectionState).toBe('reconnecting')
  })

  it('sets reconnecting on error', () => {
    const { result } = renderHook(() =>
      usePresence('Cole', 'weekly-performance'),
    )

    act(() => {
      MockEventSource.instances[0].onerror?.(new Event('error'))
    })

    expect(result.current.connectionState).toBe('reconnecting')
  })

  it('closes EventSource on unmount', () => {
    const { unmount } = renderHook(() =>
      usePresence('Cole', 'weekly-performance'),
    )

    unmount()

    expect(MockEventSource.instances[0].close).toHaveBeenCalledTimes(1)
  })

  it('closes the previous EventSource when viewerName changes', () => {
    const { rerender } = renderHook(
      ({ name }) => usePresence(name, 'weekly-performance'),
      { initialProps: { name: 'Cole' } },
    )
    const first = MockEventSource.instances[0]

    rerender({ name: 'Ada' })

    expect(first.close).toHaveBeenCalledTimes(1)
    expect(MockEventSource.instances).toHaveLength(2)
    expect(MockEventSource.instances[1].url).toContain('name=Ada')
  })

  it('closes the connection and resets viewers when viewerName is cleared', () => {
    const { result, rerender } = renderHook(
      ({ name }) => usePresence(name, 'weekly-performance'),
      { initialProps: { name: 'Cole' } },
    )

    act(() => {
      MockEventSource.instances[0].onmessage?.({
        data: JSON.stringify({ viewers: ['Cole'] }),
      } as MessageEvent)
    })

    expect(result.current.viewers).toEqual(['Cole'])

    rerender({ name: '' })

    expect(MockEventSource.instances[0].close).toHaveBeenCalledTimes(1)
    expect(result.current.viewers).toEqual([])
  })
})
