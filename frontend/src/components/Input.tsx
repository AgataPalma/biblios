interface InputProps {
    label: string
    type?: string
    value: string
    onChange: (value: string) => void
    error?: string
    placeholder?: string
}

export default function Input({ label, type = 'text', value, onChange, error, placeholder }: InputProps) {
    return (
        <div className="flex flex-col gap-1">
            <label
                className="text-sm font-medium"
                style={{ color: 'var(--color-text)' }}
            >
                {label}
            </label>
            <input
                type={type}
                value={value}
                onChange={(e) => onChange(e.target.value)}
                placeholder={placeholder}
                className="px-3 py-2 text-sm outline-none w-full"
                style={{
                    background: 'var(--input-bg)',
                    border: `1px solid ${error ? 'var(--color-error)' : 'var(--color-border)'}`,
                    borderRadius: 'var(--border-radius)',
                    color: 'var(--color-text)',
                    transition: 'var(--transition)',
                }}
            />
            {error && (
                <p className="text-xs" style={{ color: 'var(--color-error)' }}>{error}</p>
            )}
        </div>
    )
}