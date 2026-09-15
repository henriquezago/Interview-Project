import type { FormEvent } from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SignInCard } from './SignInCard'

describe('SignInCard', () => {
  it('renders the heading, name field, and submit button', () => {
    render(
      <SignInCard
        nameInput=""
        onNameChange={vi.fn()}
        onSubmit={vi.fn()}
      />,
    )

    expect(
      screen.getByRole('heading', { name: 'Open the report' }),
    ).toBeInTheDocument()
    expect(screen.getByLabelText('Your name')).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'View report' }),
    ).toBeInTheDocument()
  })

  it('calls onNameChange when the user types', async () => {
    const user = userEvent.setup()
    const onNameChange = vi.fn()

    render(
      <SignInCard
        nameInput=""
        onNameChange={onNameChange}
        onSubmit={vi.fn()}
      />,
    )

    await user.type(screen.getByLabelText('Your name'), 'Cole')

    expect(onNameChange).toHaveBeenCalled()
  })

  it('calls onSubmit when the form is submitted', async () => {
    const user = userEvent.setup()
    const onSubmit = vi.fn((event: FormEvent<HTMLFormElement>) => {
      event.preventDefault()
    })

    render(
      <SignInCard
        nameInput="Cole"
        onNameChange={vi.fn()}
        onSubmit={onSubmit}
      />,
    )

    await user.click(screen.getByRole('button', { name: 'View report' }))

    expect(onSubmit).toHaveBeenCalledTimes(1)
  })
})
