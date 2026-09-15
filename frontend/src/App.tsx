import { useState } from 'react'
import type { FormEvent } from 'react'
import { ReportPresence } from './components/ReportPresence'
import { SignInCard } from './components/SignInCard'
import { usePresence } from './hooks/usePresence'

const report = {
  id: 'weekly-performance',
  title: 'Weekly Performance Report',
}

function App() {
  const [nameInput, setNameInput] = useState('')
  const [viewerName, setViewerName] = useState('')
  const { viewers, connectionState } = usePresence(viewerName, report.id)

  const joinReport = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const name = nameInput.trim()
    if (name) {
      setViewerName(name)
    }
  }

  const leaveReport = () => {
    setViewerName('')
  }

  return (
    <main>
      {!viewerName ? (
        <SignInCard
          nameInput={nameInput}
          onNameChange={setNameInput}
          onSubmit={joinReport}
        />
      ) : (
        <ReportPresence
          title={report.title}
          viewerName={viewerName}
          viewers={viewers}
          connectionState={connectionState}
          onLeave={leaveReport}
        />
      )}
    </main>
  )
}

export default App
