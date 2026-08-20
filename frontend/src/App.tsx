import { useState } from 'react'

function App() {
  const [message, setMessage] = useState('')

  const pingBackend = () => {
    setMessage('Loading...')

    fetch('/api')
      .then((response) => {
        if (!response.ok) {
          throw new Error('Backend request failed')
        }

        return response.json()
      })
      .then((data: { message: string }) => setMessage(data.message))
      .catch(() => setMessage('Could not reach the backend.'))
  }

  return (
    <main>
      <h1>Backend response</h1>
      <button type="button" onClick={pingBackend}>
        Ping backend
      </button>
      {message && <p>{message}</p>}
    </main>
  )
}

export default App
