import type { FormEvent } from 'react'

type SignInCardProps = {
  nameInput: string
  onNameChange: (value: string) => void
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
}

export function SignInCard({
  nameInput,
  onNameChange,
  onSubmit,
}: SignInCardProps) {
  return (
    <section className="card sign-in" aria-labelledby="welcome-title">
      <p className="eyebrow">HYPE10 Reports</p>
      <h1 id="welcome-title">Open the report</h1>
      <p className="description">
        Enter your name so teammates can see that you are viewing it.
      </p>
      <form onSubmit={onSubmit}>
        <label htmlFor="name">Your name</label>
        <input
          id="name"
          maxLength={50}
          value={nameInput}
          onChange={(event) => onNameChange(event.target.value)}
          placeholder="e.g. Cole"
          autoComplete="name"
          autoFocus
          required
        />
        <button type="submit">View report</button>
      </form>
    </section>
  )
}
