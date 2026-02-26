export type ThemeId =
    | 'default-light'
    | 'woody'
    | 'nordic'
    | 'metallic'
    | 'futuristic'
    | 'post-apocalyptic'
    | 'dark-academia'
    | 'ocean'
    | 'space'

export interface Theme {
    id: ThemeId
    name: string
    description: string
    emoji: string
    fonts: {
        heading: string
        body: string
    }
    colors: {
        background: string
        surface: string
        surfaceAlt: string
        border: string
        text: string
        textMuted: string
        primary: string
        primaryHover: string
        primaryText: string
        accent: string
        error: string
        success: string
        shadow: string
    }
    effects: {
        borderRadius: string
        shadowSm: string
        shadowMd: string
        shadowLg: string
        transition: string
        inputBg: string
        glassBg: string
        glassBlur: string
    }
}

export const themes: Record<ThemeId, Theme> = {
    'default-light': {
        id: 'default-light',
        name: 'Default Light',
        description: 'Clean and simple',
        emoji: '☀️',
        fonts: {
            heading: "'Inter', sans-serif",
            body: "'Inter', sans-serif",
        },
        colors: {
            background: '#f9fafb',
            surface: '#ffffff',
            surfaceAlt: '#f3f4f6',
            border: '#e5e7eb',
            text: '#111827',
            textMuted: '#6b7280',
            primary: '#2563eb',
            primaryHover: '#1d4ed8',
            primaryText: '#ffffff',
            accent: '#3b82f6',
            error: '#ef4444',
            success: '#22c55e',
            shadow: 'rgba(0,0,0,0.1)',
        },
        effects: {
            borderRadius: '8px',
            shadowSm: '0 1px 3px rgba(0,0,0,0.1)',
            shadowMd: '0 4px 6px rgba(0,0,0,0.1)',
            shadowLg: '0 10px 25px rgba(0,0,0,0.1)',
            transition: 'all 0.2s ease',
            inputBg: '#ffffff',
            glassBg: 'rgba(255,255,255,0.8)',
            glassBlur: 'blur(10px)',
        },
    },

    'woody': {
        id: 'woody',
        name: 'Woody Cabin',
        description: 'Warm wood and fireplace vibes',
        emoji: '🪵',
        fonts: {
            heading: "'Playfair Display', serif",
            body: "'Lato', sans-serif",
        },
        colors: {
            background: '#2c1810',
            surface: '#3d2314',
            surfaceAlt: '#4a2c1a',
            border: '#6b3a23',
            text: '#f5e6d3',
            textMuted: '#c4a882',
            primary: '#c17f3a',
            primaryHover: '#a66830',
            primaryText: '#ffffff',
            accent: '#e8a045',
            error: '#e05252',
            success: '#6aab6a',
            shadow: 'rgba(0,0,0,0.4)',
        },
        effects: {
            borderRadius: '4px',
            shadowSm: '0 2px 4px rgba(0,0,0,0.3)',
            shadowMd: '0 4px 12px rgba(0,0,0,0.4)',
            shadowLg: '0 8px 30px rgba(0,0,0,0.5)',
            transition: 'all 0.3s ease',
            inputBg: '#3d2314',
            glassBg: 'rgba(61,35,20,0.85)',
            glassBlur: 'blur(8px)',
        },
    },

    'nordic': {
        id: 'nordic',
        name: 'Nordic',
        description: 'Clean Scandinavian minimalism',
        emoji: '❄️',
        fonts: {
            heading: "'Raleway', sans-serif",
            body: "'Open Sans', sans-serif",
        },
        colors: {
            background: '#eceff4',
            surface: '#ffffff',
            surfaceAlt: '#e5e9f0',
            border: '#d8dee9',
            text: '#2e3440',
            textMuted: '#4c566a',
            primary: '#5e81ac',
            primaryHover: '#4c6f9a',
            primaryText: '#ffffff',
            accent: '#88c0d0',
            error: '#bf616a',
            success: '#a3be8c',
            shadow: 'rgba(46,52,64,0.1)',
        },
        effects: {
            borderRadius: '2px',
            shadowSm: '0 1px 3px rgba(46,52,64,0.08)',
            shadowMd: '0 3px 8px rgba(46,52,64,0.1)',
            shadowLg: '0 8px 20px rgba(46,52,64,0.12)',
            transition: 'all 0.15s ease',
            inputBg: '#ffffff',
            glassBg: 'rgba(255,255,255,0.9)',
            glassBlur: 'blur(12px)',
        },
    },

    'metallic': {
        id: 'metallic',
        name: 'Metallic',
        description: 'Industrial steel and chrome',
        emoji: '⚙️',
        fonts: {
            heading: "'Oswald', sans-serif",
            body: "'Roboto', sans-serif",
        },
        colors: {
            background: '#1a1a1a',
            surface: '#242424',
            surfaceAlt: '#2e2e2e',
            border: '#444444',
            text: '#d4d4d4',
            textMuted: '#888888',
            primary: '#9e9e9e',
            primaryHover: '#b5b5b5',
            primaryText: '#000000',
            accent: '#c0c0c0',
            error: '#cf4444',
            success: '#4caf50',
            shadow: 'rgba(0,0,0,0.5)',
        },
        effects: {
            borderRadius: '2px',
            shadowSm: '0 1px 3px rgba(0,0,0,0.4), inset 0 1px 0 rgba(255,255,255,0.05)',
            shadowMd: '0 4px 8px rgba(0,0,0,0.5), inset 0 1px 0 rgba(255,255,255,0.05)',
            shadowLg: '0 8px 24px rgba(0,0,0,0.6)',
            transition: 'all 0.2s ease',
            inputBg: '#1a1a1a',
            glassBg: 'rgba(36,36,36,0.9)',
            glassBlur: 'blur(8px)',
        },
    },

    'futuristic': {
        id: 'futuristic',
        name: 'Futuristic',
        description: 'Cyberpunk neon on dark',
        emoji: '🤖',
        fonts: {
            heading: "'Orbitron', sans-serif",
            body: "'Exo 2', sans-serif",
        },
        colors: {
            background: '#050510',
            surface: '#0a0a1f',
            surfaceAlt: '#0f0f2d',
            border: '#1a1a4a',
            text: '#e0e0ff',
            textMuted: '#8080bb',
            primary: '#00f0ff',
            primaryHover: '#00d0dd',
            primaryText: '#000020',
            accent: '#ff00ff',
            error: '#ff3366',
            success: '#00ff88',
            shadow: 'rgba(0,240,255,0.2)',
        },
        effects: {
            borderRadius: '0px',
            shadowSm: '0 0 8px rgba(0,240,255,0.2)',
            shadowMd: '0 0 16px rgba(0,240,255,0.3)',
            shadowLg: '0 0 32px rgba(0,240,255,0.4)',
            transition: 'all 0.1s ease',
            inputBg: '#0a0a1f',
            glassBg: 'rgba(10,10,31,0.85)',
            glassBlur: 'blur(12px)',
        },
    },

    'post-apocalyptic': {
        id: 'post-apocalyptic',
        name: 'Post-Apocalyptic',
        description: 'The Last of Us survival mode',
        emoji: '🍄',
        fonts: {
            heading: "'Bebas Neue', sans-serif",
            body: "'Special Elite', serif",
        },
        colors: {
            background: '#1a1208',
            surface: '#241a0a',
            surfaceAlt: '#2e2010',
            border: '#4a3520',
            text: '#c8b89a',
            textMuted: '#7a6a52',
            primary: '#8b6914',
            primaryHover: '#a07a18',
            primaryText: '#f0e0c0',
            accent: '#c8960a',
            error: '#8b2020',
            success: '#4a6a20',
            shadow: 'rgba(0,0,0,0.6)',
        },
        effects: {
            borderRadius: '2px',
            shadowSm: '0 2px 4px rgba(0,0,0,0.5)',
            shadowMd: '0 4px 12px rgba(0,0,0,0.6)',
            shadowLg: '0 8px 24px rgba(0,0,0,0.7)',
            transition: 'all 0.3s ease',
            inputBg: '#1a1208',
            glassBg: 'rgba(36,26,10,0.9)',
            glassBlur: 'blur(6px)',
        },
    },

    'dark-academia': {
        id: 'dark-academia',
        name: 'Dark Academia',
        description: 'Old books and candlelight',
        emoji: '📚',
        fonts: {
            heading: "'Cormorant Garamond', serif",
            body: "'EB Garamond', serif",
        },
        colors: {
            background: '#1c1712',
            surface: '#252015',
            surfaceAlt: '#2e2818',
            border: '#4a3f2a',
            text: '#e8dcc8',
            textMuted: '#a09070',
            primary: '#8b6914',
            primaryHover: '#7a5a10',
            primaryText: '#f0e0c0',
            accent: '#c8a050',
            error: '#8b3030',
            success: '#4a6a30',
            shadow: 'rgba(0,0,0,0.5)',
        },
        effects: {
            borderRadius: '2px',
            shadowSm: '0 2px 6px rgba(0,0,0,0.4)',
            shadowMd: '0 4px 12px rgba(0,0,0,0.5)',
            shadowLg: '0 8px 30px rgba(0,0,0,0.6)',
            transition: 'all 0.3s ease',
            inputBg: '#1c1712',
            glassBg: 'rgba(37,32,21,0.9)',
            glassBlur: 'blur(8px)',
        },
    },

    'ocean': {
        id: 'ocean',
        name: 'Ocean',
        description: 'Deep blue calm waters',
        emoji: '🌊',
        fonts: {
            heading: "'Nunito', sans-serif",
            body: "'Source Sans 3', sans-serif",
        },
        colors: {
            background: '#0a1628',
            surface: '#0f2040',
            surfaceAlt: '#142850',
            border: '#1e3a5f',
            text: '#c8e0f0',
            textMuted: '#6890b0',
            primary: '#0080c0',
            primaryHover: '#0070a8',
            primaryText: '#ffffff',
            accent: '#00c8e0',
            error: '#e05050',
            success: '#20b080',
            shadow: 'rgba(0,0,0,0.4)',
        },
        effects: {
            borderRadius: '12px',
            shadowSm: '0 2px 8px rgba(0,128,192,0.15)',
            shadowMd: '0 4px 16px rgba(0,128,192,0.2)',
            shadowLg: '0 8px 32px rgba(0,128,192,0.25)',
            transition: 'all 0.3s ease',
            inputBg: '#0f2040',
            glassBg: 'rgba(15,32,64,0.85)',
            glassBlur: 'blur(12px)',
        },
    },

    'space': {
        id: 'space',
        name: 'Space',
        description: 'Cosmos and starfields',
        emoji: '🚀',
        fonts: {
            heading: "'Space Grotesk', sans-serif",
            body: "'Space Grotesk', sans-serif",
        },
        colors: {
            background: '#030308',
            surface: '#080818',
            surfaceAlt: '#0d0d25',
            border: '#1a1a3a',
            text: '#e0e8ff',
            textMuted: '#6070a0',
            primary: '#6040c0',
            primaryHover: '#7050d0',
            primaryText: '#ffffff',
            accent: '#a060ff',
            error: '#ff4060',
            success: '#40c080',
            shadow: 'rgba(96,64,192,0.3)',
        },
        effects: {
            borderRadius: '6px',
            shadowSm: '0 0 8px rgba(96,64,192,0.2)',
            shadowMd: '0 0 20px rgba(96,64,192,0.3)',
            shadowLg: '0 0 40px rgba(96,64,192,0.4)',
            transition: 'all 0.2s ease',
            inputBg: '#080818',
            glassBg: 'rgba(8,8,24,0.85)',
            glassBlur: 'blur(12px)',
        },
    },
}