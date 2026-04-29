/**
 * Returns a human-readable relative date string for a given ISO 8601 timestamp.
 *
 * Examples:
 *   < 1 minute ago  → "just now"
 *   < 60 minutes    → "5m ago"
 *   < 24 hours      → "3h ago"
 *   < 7 days        → "2d ago"
 *   older           → "12 Jan 2024"
 */
export function relativeDate(iso: string): string {
    const diff = Date.now() - new Date(iso).getTime()
    const mins = Math.floor(diff / 60_000)

    if (mins < 1)  return 'just now'
    if (mins < 60) return `${mins}m ago`

    const hrs = Math.floor(mins / 60)
    if (hrs < 24)  return `${hrs}h ago`

    const days = Math.floor(hrs / 24)
    if (days < 7)  return `${days}d ago`

    return new Date(iso).toLocaleDateString('en-GB', {
        day: 'numeric',
        month: 'short',
        year: 'numeric',
    })
}
