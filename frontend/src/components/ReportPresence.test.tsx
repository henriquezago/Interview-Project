import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ReportPresence } from './ReportPresence'

const baseProps = {
  title: 'Weekly Performance Report',
  viewerName: 'Cole',
  viewers: [] as string[],
  connectionState: 'connecting' as const,
  onLeave: vi.fn(),
}

describe('ReportPresence', () => {
  it('renders the report title and calls onLeave when Leave is clicked', async () => {
    const user = userEvent.setup()
    const onLeave = vi.fn()

    render(<ReportPresence {...baseProps} onLeave={onLeave} />)

    expect(
      screen.getByRole('heading', { name: 'Weekly Performance Report' }),
    ).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Leave' }))

    expect(onLeave).toHaveBeenCalledTimes(1)
  })

  it('shows Live when connected and the raw state otherwise', () => {
    const { rerender } = render(
      <ReportPresence {...baseProps} connectionState="connected" />,
    )

    expect(screen.getByRole('status')).toHaveTextContent('Live')

    rerender(<ReportPresence {...baseProps} connectionState="reconnecting" />)

    expect(screen.getByRole('status')).toHaveTextContent('reconnecting')
  })

  it('lists viewers and marks the current viewer as You', () => {
    render(
      <ReportPresence
        {...baseProps}
        viewers={['Cole', 'Ada']}
        connectionState="connected"
      />,
    )

    expect(screen.getByText('Cole')).toBeInTheDocument()
    expect(screen.getByText('Ada')).toBeInTheDocument()
    expect(screen.getByText('You')).toBeInTheDocument()
  })

  it('shows an empty state when there are no viewers', () => {
    render(<ReportPresence {...baseProps} />)

    expect(
      screen.getByText('Waiting for viewer information…'),
    ).toBeInTheDocument()
  })
})
