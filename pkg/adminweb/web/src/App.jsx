import { Routes, Route, Navigate } from 'react-router-dom'
import ProtectedRoute from './components/ProtectedRoute'
import Layout from './components/Layout'
import Login from './pages/Login'
import Tokens from './pages/Tokens'
import Devices from './pages/Devices'
import Webhooks from './pages/Webhooks'

function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route index element={<Navigate to="/tokens" replace />} />
        <Route path="tokens" element={<Tokens />} />
        <Route path="devices" element={<Devices />} />
        <Route path="webhooks" element={<Webhooks />} />
      </Route>
    </Routes>
  )
}

export default App
