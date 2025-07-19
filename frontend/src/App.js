import React, { useState, useEffect } from 'react'
import axios from 'axios'
import './App.css'

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080'

function App() {
  const [users, setUsers] = useState([])
  const [newUser, setNewUser] = useState({ name: '', email: '' })
  const [loading, setLoading] = useState(false)
  const [stressTestResult, setStressTestResult] = useState('')

  useEffect(() => {
    fetchUsers()
  }, [])

  const fetchUsers = async () => {
    try {
      const response = await axios.get(`${API_URL}/api/users`)
      setUsers(response.data || [])
    } catch (error) {
      console.error('Error fetching users:', error)
    }
  }

  const createUser = async (e) => {
    e.preventDefault()
    try {
      await axios.post(`${API_URL}/api/users`, newUser)
      setNewUser({ name: '', email: '' })
      fetchUsers()
    } catch (error) {
      console.error('Error creating user:', error)
    }
  }

  const runStressTest = async () => {
    setLoading(true)
    setStressTestResult('Running stress test...')
    try {
      const response = await axios.get(`${API_URL}/api/stress`)
      setStressTestResult(
        'Stress test completed: ' + JSON.stringify(response.data)
      )
    } catch (error) {
      setStressTestResult('Stress test failed: ' + error.message)
    }
    setLoading(false)
  }

  return (
    <div className='App'>
      <h1>Kubernetes Auto-Scaling Demo</h1>

      <div className='section'>
        <h2>Create User</h2>
        <form onSubmit={createUser}>
          <input
            type='text'
            placeholder='Name'
            value={newUser.name}
            onChange={(e) => setNewUser({ ...newUser, name: e.target.value })}
            required
          />
          <input
            type='email'
            placeholder='Email'
            value={newUser.email}
            onChange={(e) => setNewUser({ ...newUser, email: e.target.value })}
            required
          />
          <button type='submit'>Create User</button>
        </form>
      </div>

      <div className='section'>
        <h2>Users</h2>
        <ul>
          {users.map((user) => (
            <li key={user.id}>
              {user.name} - {user.email}
            </li>
          ))}
        </ul>
      </div>

      <div className='section'>
        <h2>Load Testing</h2>
        <button onClick={runStressTest} disabled={loading}>
          {loading ? 'Running...' : 'Run Stress Test'}
        </button>
        {stressTestResult && <p>{stressTestResult}</p>}
      </div>
    </div>
  )
}

export default App