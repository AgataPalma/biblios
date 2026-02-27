import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from './context/AuthContext'
import { Spinner } from './components'
import Navbar from './components/Navbar'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import DashboardPage from './pages/DashboardPage'
import BooksPage from './pages/BooksPage'
import BookDetailPage from './pages/BookDetailPage'
import LibraryPage from './pages/LibraryPage'
import ModerationPage from './pages/ModerationPage'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
    const { isAuthenticated, isLoading } = useAuth()

    if (isLoading) {
        return (
            <div style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                minHeight: '100vh',
            }}>
                <Spinner size="lg" label="Loading..." />
            </div>
        )
    }

    if (!isAuthenticated) {
        return <Navigate to="/login" replace />
    }

    return (
        <>
            <Navbar />
            <main style={{ paddingTop: '56px' }}>
                {children}
            </main>
        </>
    )
}

export default function App() {
    return (
        <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/" element={
                <ProtectedRoute>
                    <DashboardPage />
                </ProtectedRoute>
            } />
            <Route path="/books" element={
                <ProtectedRoute>
                    <BooksPage />
                </ProtectedRoute>
            } />
            <Route path="/books/:id" element={
                <ProtectedRoute>
                    <BookDetailPage />
                </ProtectedRoute>
            } />
            <Route path="/library" element={
                <ProtectedRoute>
                    <LibraryPage />
                </ProtectedRoute>
            } />
            <Route path="/moderation" element={
                <ProtectedRoute>
                    <ModerationPage />
                </ProtectedRoute>
            } />
        </Routes>
    )
}