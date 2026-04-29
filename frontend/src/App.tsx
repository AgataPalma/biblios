import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from './context/AuthContext'
import { Spinner } from './components'
import Sidebar from './components/Sidebar'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import DashboardPage from './pages/DashboardPage'
import BooksPage from './pages/BooksPage'
import BookDetailPage from './pages/BookDetailPage'
import LibraryPage from './pages/LibraryPage'
import ModerationPage from './pages/ModerationPage'
import AddBookPage from './pages/AddBookPage'
import ProfilePage from './pages/ProfilePage'
import NotFoundPage from './pages/NotFoundPage'
import SettingsPage from './pages/SettingsPage'
import NotificationsPage from './pages/NotificationsPage'
import ShelvesPage from './pages/ShelvesPage'
import ShelfDetailPage from './pages/ShelfDetailPage'
import LibrariesPage from './pages/LibrariesPage'
import LibraryDetailPage from './pages/LibraryDetailPage'
import InvitationAcceptPage from './pages/InvitationAcceptPage'
import SeriesPage from './pages/SeriesPage'
import SeriesDetailPage from './pages/SeriesDetailPage'
import ReadingPage from './pages/ReadingPage'
import BookListsPage from './pages/BookListsPage'
import BookListDetailPage from './pages/BookListDetailPage'
import CollectionDetailPage from './pages/CollectionDetailPage'

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
        <div style={{ display: 'flex', minHeight: '100vh' }}>
            <Sidebar />
            {/* paddingLeft matches --sidebar-w which Sidebar sets on <html> */}
            <main style={{
                flex: 1,
                minWidth: 0,
                paddingLeft: 'var(--sidebar-w, 220px)',
                transition: 'padding-left 0.22s ease',
            }}>
                {children}
            </main>
        </div>
    )
}

export default function App() {
    return (
        <Routes>
            <Route path="/login"    element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/" element={
                <ProtectedRoute><DashboardPage /></ProtectedRoute>
            } />
            <Route path="/profile" element={
                <ProtectedRoute><ProfilePage /></ProtectedRoute>
            } />
            <Route path="/books" element={
                <ProtectedRoute><BooksPage /></ProtectedRoute>
            } />
            <Route path="/books/add" element={
                <ProtectedRoute><AddBookPage /></ProtectedRoute>
            } />
            <Route path="/books/:id" element={
                <ProtectedRoute><BookDetailPage /></ProtectedRoute>
            } />
            <Route path="/library" element={
                <ProtectedRoute><LibraryPage /></ProtectedRoute>
            } />
            <Route path="/moderation" element={
                <ProtectedRoute><ModerationPage /></ProtectedRoute>
            } />
            <Route path="/settings" element={
                <ProtectedRoute><SettingsPage /></ProtectedRoute>
            } />
            <Route path="/notifications" element={
                <ProtectedRoute><NotificationsPage /></ProtectedRoute>
            } />
            <Route path="/shelves" element={
                <ProtectedRoute><ShelvesPage /></ProtectedRoute>
            } />
            <Route path="/shelves/:id" element={
                <ProtectedRoute><ShelfDetailPage /></ProtectedRoute>
            } />
            <Route path="/libraries/invitations/:token/accept" element={
                <ProtectedRoute><InvitationAcceptPage /></ProtectedRoute>
            } />
            <Route path="/libraries" element={
                <ProtectedRoute><LibrariesPage /></ProtectedRoute>
            } />
            <Route path="/libraries/:id" element={
                <ProtectedRoute><LibraryDetailPage /></ProtectedRoute>
            } />
            <Route path="/series" element={
                <ProtectedRoute><SeriesPage /></ProtectedRoute>
            } />
            <Route path="/series/:id" element={
                <ProtectedRoute><SeriesDetailPage /></ProtectedRoute>
            } />
            <Route path="/reading" element={
                <ProtectedRoute><ReadingPage /></ProtectedRoute>
            } />
            <Route path="/lists" element={
                <ProtectedRoute><BookListsPage /></ProtectedRoute>
            } />
            <Route path="/lists/:id" element={
                <ProtectedRoute><BookListDetailPage /></ProtectedRoute>
            } />
            <Route path="/collections/:id" element={
                <ProtectedRoute><CollectionDetailPage /></ProtectedRoute>
            } />
            <Route path="*" element={<NotFoundPage />} />
        </Routes>
    )
}