import { useEffect } from 'react'
import type { ReactNode } from 'react'
import Button from './Button'

interface ModalProps {
    isOpen: boolean
    onClose: () => void
    title: string
    children: ReactNode
    confirmLabel?: string
    onConfirm?: () => void
    confirmVariant?: 'primary' | 'danger'
    isLoading?: boolean
    size?: 'sm' | 'md' | 'lg'
}

export default function Modal({
                                  isOpen,
                                  onClose,
                                  title,
                                  children,
                                  confirmLabel,
                                  onConfirm,
                                  confirmVariant = 'primary',
                                  isLoading = false,
                                  size = 'md',
                              }: ModalProps) {
    // Close on Escape key
    useEffect(() => {
        function handleKey(e: KeyboardEvent) {
            if (e.key === 'Escape') onClose()
        }
        if (isOpen) document.addEventListener('keydown', handleKey)
        return () => document.removeEventListener('keydown', handleKey)
    }, [isOpen, onClose])

    // Prevent body scroll when open
    useEffect(() => {
        if (isOpen) document.body.style.overflow = 'hidden'
        else document.body.style.overflow = ''
        return () => { document.body.style.overflow = '' }
    }, [isOpen])

    if (!isOpen) return null

    const widths = { sm: '380px', md: '520px', lg: '700px' }

    return (
        <div
            style={{
                position: 'fixed',
                inset: 0,
                zIndex: 1000,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                padding: '16px',
                background: 'rgba(0,0,0,0.6)',
                backdropFilter: 'blur(4px)',
            }}
            onClick={e => { if (e.target === e.currentTarget) onClose() }}
        >
            <div
                style={{
                    background: 'var(--color-surface)',
                    border: '1px solid var(--color-border)',
                    borderRadius: 'var(--border-radius)',
                    boxShadow: 'var(--shadow-lg)',
                    width: '100%',
                    maxWidth: widths[size],
                    maxHeight: '90vh',
                    display: 'flex',
                    flexDirection: 'column',
                    fontFamily: 'var(--font-body)',
                }}
            >
                {/* Header */}
                <div
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        padding: '16px 20px',
                        borderBottom: '1px solid var(--color-border)',
                    }}
                >
                    <h2
                        style={{
                            margin: 0,
                            fontSize: '16px',
                            fontWeight: 600,
                            color: 'var(--color-text)',
                            fontFamily: 'var(--font-heading)',
                        }}
                    >
                        {title}
                    </h2>
                    <button
                        onClick={onClose}
                        style={{
                            background: 'none',
                            border: 'none',
                            cursor: 'pointer',
                            color: 'var(--color-text-muted)',
                            fontSize: '20px',
                            lineHeight: 1,
                            padding: '4px',
                            borderRadius: 'var(--border-radius)',
                        }}
                    >
                        ×
                    </button>
                </div>

                {/* Body */}
                <div
                    style={{
                        padding: '20px',
                        overflowY: 'auto',
                        color: 'var(--color-text)',
                        flex: 1,
                    }}
                >
                    {children}
                </div>

                {/* Footer */}
                {(confirmLabel || onConfirm) && (
                    <div
                        style={{
                            display: 'flex',
                            justifyContent: 'flex-end',
                            gap: '8px',
                            padding: '16px 20px',
                            borderTop: '1px solid var(--color-border)',
                        }}
                    >
                        <Button label="Cancel" variant="secondary" onClick={onClose} />
                        {confirmLabel && onConfirm && (
                            <Button
                                label={confirmLabel}
                                variant={confirmVariant}
                                onClick={onConfirm}
                                isLoading={isLoading}
                            />
                        )}
                    </div>
                )}
            </div>
        </div>
    )
}