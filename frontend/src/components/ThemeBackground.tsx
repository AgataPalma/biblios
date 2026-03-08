import { useMemo } from 'react'
import { useTheme } from '../context/ThemeContext'
import type { ThemeId } from '../themes/themes'

// ─── Theme configurations ────────────────────────────────────────────────────

interface BookcaseTheme {
    // Shelf material
    shelfColor: string
    shelfHighlight: string
    shelfShadow: string
    shelfEdge: string
    // Frame/case material
    frameColor: string
    frameHighlight: string
    frameShadow: string
    // Background (wall behind books)
    wallColor: string
    wallPattern?: string
    // Ambient light/overlay
    ambientColor: string
    ambientOpacity: number
    // Book colour palettes
    bookPalettes: string[][]
    // Spine texture style
    spineStyle: 'leather' | 'cloth' | 'metal' | 'worn' | 'glossy' | 'matte' | 'cracked' | 'circuit' | 'parchment'
    // Overall darkness (0-1)
    darkness: number
    // Grain/noise overlay opacity
    grainOpacity: number
}

const bookcaseThemes: Record<ThemeId, BookcaseTheme> = {
    'default-light': {
        shelfColor: '#d4b896',
        shelfHighlight: '#e8d0b4',
        shelfShadow: '#a07850',
        shelfEdge: '#8b6040',
        frameColor: '#c4a882',
        frameHighlight: '#dcc09a',
        frameShadow: '#8a6030',
        wallColor: '#f5efe6',
        ambientColor: '#fff8f0',
        ambientOpacity: 0.15,
        bookPalettes: [
            ['#2563eb', '#1d4ed8', '#3b82f6'],
            ['#16a34a', '#15803d', '#22c55e'],
            ['#dc2626', '#b91c1c', '#ef4444'],
            ['#9333ea', '#7e22ce', '#a855f7'],
            ['#ea580c', '#c2410c', '#f97316'],
            ['#0891b2', '#0e7490', '#06b6d4'],
            ['#4f46e5', '#4338ca', '#6366f1'],
            ['#d97706', '#b45309', '#f59e0b'],
        ],
        spineStyle: 'cloth',
        darkness: 0.1,
        grainOpacity: 0.03,
    },
    'woody': {
        shelfColor: '#5c3317',
        shelfHighlight: '#7a4520',
        shelfShadow: '#2d1608',
        shelfEdge: '#1a0c04',
        frameColor: '#4a2810',
        frameHighlight: '#6b3c18',
        frameShadow: '#1a0a04',
        wallColor: '#1a0c05',
        ambientColor: '#ff8c00',
        ambientOpacity: 0.08,
        bookPalettes: [
            ['#8b4513', '#6b3410', '#a0522d'],
            ['#556b2f', '#3d4f22', '#6b8540'],
            ['#8b6914', '#6b5010', '#a07820'],
            ['#4a3520', '#2d2010', '#5c4428'],
            ['#8b3010', '#6b2008', '#a03818'],
            ['#3d5555', '#2d4040', '#4d6868'],
            ['#6b4020', '#4a2c10', '#8b5030'],
            ['#556040', '#404830', '#6b7850'],
        ],
        spineStyle: 'leather',
        darkness: 0.7,
        grainOpacity: 0.12,
    },
    'nordic': {
        shelfColor: '#c8d8e8',
        shelfHighlight: '#dce8f4',
        shelfShadow: '#8fa8c0',
        shelfEdge: '#6888a0',
        frameColor: '#b8ccd8',
        frameHighlight: '#ccdce8',
        frameShadow: '#7898b0',
        wallColor: '#eef2f8',
        ambientColor: '#88c0d0',
        ambientOpacity: 0.12,
        bookPalettes: [
            ['#5e81ac', '#4c6f9a', '#7090c0'],
            ['#88c0d0', '#70aac0', '#a0d0e0'],
            ['#a3be8c', '#8aaa74', '#b8d0a0'],
            ['#bf616a', '#a85060', '#d07080'],
            ['#ebcb8b', '#d8b870', '#f0d8a0'],
            ['#b48ead', '#a07898', '#c8a0c0'],
            ['#d8dee9', '#c0ccdc', '#e8eef8'],
            ['#4c566a', '#3c4558', '#5c6880'],
        ],
        spineStyle: 'matte',
        darkness: 0.05,
        grainOpacity: 0.02,
    },
    'metallic': {
        shelfColor: '#3a3a3a',
        shelfHighlight: '#585858',
        shelfShadow: '#181818',
        shelfEdge: '#101010',
        frameColor: '#2a2a2a',
        frameHighlight: '#484848',
        frameShadow: '#080808',
        wallColor: '#141414',
        ambientColor: '#c0c0c0',
        ambientOpacity: 0.06,
        bookPalettes: [
            ['#606060', '#484848', '#787878'],
            ['#808080', '#686868', '#989898'],
            ['#404040', '#303030', '#505050'],
            ['#5a5a6a', '#484858', '#6a6a7a'],
            ['#3a4040', '#2a3030', '#4a5050'],
            ['#505058', '#404048', '#606068'],
            ['#383838', '#282828', '#484848'],
            ['#4a4a4a', '#383838', '#5a5a5a'],
        ],
        spineStyle: 'metal',
        darkness: 0.8,
        grainOpacity: 0.08,
    },
    'futuristic': {
        shelfColor: '#0a1a2a',
        shelfHighlight: '#0f2535',
        shelfShadow: '#020810',
        shelfEdge: '#00f0ff',
        frameColor: '#050f1a',
        frameHighlight: '#0a1a28',
        frameShadow: '#010508',
        wallColor: '#020510',
        ambientColor: '#00f0ff',
        ambientOpacity: 0.1,
        bookPalettes: [
            ['#00f0ff', '#00c8d8', '#40ffff'],
            ['#ff00ff', '#d800d8', '#ff40ff'],
            ['#0050ff', '#0030d8', '#2070ff'],
            ['#00ff88', '#00d870', '#40ffa0'],
            ['#ff3366', '#d81848', '#ff5080'],
            ['#8800ff', '#6800d8', '#a020ff'],
            ['#00c8ff', '#00a8d8', '#20d8ff'],
            ['#ff8800', '#d87000', '#ffa020'],
        ],
        spineStyle: 'circuit',
        darkness: 0.9,
        grainOpacity: 0.04,
    },
    'post-apocalyptic': {
        shelfColor: '#3d2e10',
        shelfHighlight: '#4a3818',
        shelfShadow: '#1a1206',
        shelfEdge: '#0d0904',
        frameColor: '#2d2008',
        frameHighlight: '#3a2c10',
        frameShadow: '#100c04',
        wallColor: '#0d0904',
        ambientColor: '#c88010',
        ambientOpacity: 0.08,
        bookPalettes: [
            ['#5a3a10', '#3d2808', '#6b4818'],
            ['#3a3020', '#282010', '#4a3c28'],
            ['#4a3808', '#332804', '#5a4810'],
            ['#5a2808', '#3d1c04', '#6b3810'],
            ['#3d4020', '#2a2c14', '#4d5028'],
            ['#484030', '#322c20', '#585040'],
            ['#3a2010', '#281408', '#4a2c18'],
            ['#504030', '#382e20', '#604e3c'],
        ],
        spineStyle: 'cracked',
        darkness: 0.85,
        grainOpacity: 0.18,
    },
    'dark-academia': {
        shelfColor: '#3d2e18',
        shelfHighlight: '#4d3c20',
        shelfShadow: '#1a1208',
        shelfEdge: '#0d0904',
        frameColor: '#2d2010',
        frameHighlight: '#3c2c18',
        frameShadow: '#100c04',
        wallColor: '#100d08',
        ambientColor: '#c8a050',
        ambientOpacity: 0.1,
        bookPalettes: [
            ['#8b6914', '#6b5010', '#a07820'],
            ['#6b3a14', '#4a2808', '#8b4c20'],
            ['#4a5030', '#343820', '#5c6438'],
            ['#6b4a20', '#4a3214', '#8b6030'],
            ['#503828', '#382818', '#644838'],
            ['#3a5040', '#283828', '#4c6450'],
            ['#7a6030', '#584420', '#8c7040'],
            ['#5a3030', '#3c2020', '#6e4040'],
        ],
        spineStyle: 'leather',
        darkness: 0.75,
        grainOpacity: 0.14,
    },
    'ocean': {
        shelfColor: '#0a2040',
        shelfHighlight: '#0f2c55',
        shelfShadow: '#040f20',
        shelfEdge: '#020810',
        frameColor: '#061830',
        frameHighlight: '#0a2040',
        frameShadow: '#020a18',
        wallColor: '#030d20',
        ambientColor: '#0080c0',
        ambientOpacity: 0.12,
        bookPalettes: [
            ['#0080c0', '#0060a0', '#20a0d8'],
            ['#00c8e0', '#00a8c0', '#20d8f0'],
            ['#0050a0', '#003880', '#1068b8'],
            ['#006880', '#004c60', '#1080a0'],
            ['#4080c0', '#3060a0', '#5898d8'],
            ['#0040a0', '#002880', '#1058b8'],
            ['#008080', '#006060', '#10a0a0'],
            ['#2050a0', '#103880', '#3068b8'],
        ],
        spineStyle: 'glossy',
        darkness: 0.85,
        grainOpacity: 0.05,
    },
    'space': {
        shelfColor: '#0d0820',
        shelfHighlight: '#140c2c',
        shelfShadow: '#04030a',
        shelfEdge: '#6040c0',
        frameColor: '#08051a',
        frameHighlight: '#0f0824',
        frameShadow: '#020108',
        wallColor: '#030108',
        ambientColor: '#6040c0',
        ambientOpacity: 0.1,
        bookPalettes: [
            ['#6040c0', '#4828a0', '#7858d8'],
            ['#a060ff', '#8040e0', '#b878ff'],
            ['#4060c0', '#2848a0', '#5878d8'],
            ['#c040a0', '#a02880', '#d858b8'],
            ['#4080c0', '#2860a0', '#5898d8'],
            ['#8040c0', '#6028a0', '#9858d8'],
            ['#2040a0', '#102880', '#3058b8'],
            ['#c06040', '#a04828', '#d87858'],
        ],
        spineStyle: 'glossy',
        darkness: 0.9,
        grainOpacity: 0.06,
    },

    'candlelight': {
        shelfColor: '#2e2620',
        shelfHighlight: '#3d3328',
        shelfShadow: '#141008',
        shelfEdge: '#0d0a06',
        frameColor: '#231e18',
        frameHighlight: '#302820',
        frameShadow: '#0a0806',
        wallColor: '#100e0a',
        ambientColor: '#c9a96e',
        ambientOpacity: 0.12,
        bookPalettes: [
            ['#7c6f5e', '#5c5040', '#9c8e7a'],
            ['#5a7a8a', '#3a5a6a', '#7a9aaa'],
            ['#7a5a6a', '#5a3a4a', '#9a7a8a'],
            ['#6a7a5a', '#4a5a3a', '#8a9a7a'],
            ['#8a6a5a', '#6a4a3a', '#aa8a7a'],
            ['#6a5a8a', '#4a3a6a', '#8a7aaa'],
            ['#8a7a5a', '#6a5a3a', '#aaa07a'],
            ['#5a6a7a', '#3a4a5a', '#7a8a9a'],
        ],
        spineStyle: 'parchment',
        darkness: 0.78,
        grainOpacity: 0.14,
    },
}

// ─── SVG generation helpers ──────────────────────────────────────────────────

function seededRandom(seed: number) {
    let s = seed
    return () => {
        s = (s * 1664525 + 1013904223) & 0xffffffff
        return (s >>> 0) / 0xffffffff
    }
}

function lerp(a: string, b: string, t: number): string {
    const parse = (c: string) => [
        parseInt(c.slice(1, 3), 16),
        parseInt(c.slice(3, 5), 16),
        parseInt(c.slice(5, 7), 16),
    ]
    const [ar, ag, ab] = parse(a)
    const [br, bg, bb] = parse(b)
    const r = Math.round(ar + (br - ar) * t)
    const g = Math.round(ag + (bg - ag) * t)
    const bl2 = Math.round(ab + (bb - ab) * t)
    return `rgb(${r},${g},${bl2})`
}

interface BookSpec {
    x: number
    width: number
    height: number
    colors: string[]
    hasTitle: boolean
    isLying: boolean
    lyingBooks?: BookSpec[]
}

function generateShelf(
    rng: () => number,
    shelfWidth: number,
    _shelfHeight: number,
    theme: BookcaseTheme,
    gapChance: number
): BookSpec[] {
    const books: BookSpec[] = []
    let x = 8
    const maxX = shelfWidth - 8
    const frameLeft = 12

    while (x < maxX - 15) {
        // Occasional gap
        if (rng() < gapChance && x > frameLeft + 20) {
            // Sometimes a lying book in the gap
            if (rng() < 0.4) {
                const gapW = 40 + rng() * 60
                const lyingH = 12 + rng() * 10
                const palette = theme.bookPalettes[Math.floor(rng() * theme.bookPalettes.length)]
                books.push({
                    x,
                    width: gapW,
                    height: lyingH,
                    colors: palette,
                    hasTitle: false,
                    isLying: true,
                })
                x += gapW + 4
            } else {
                x += 20 + rng() * 40
            }
            continue
        }

        const bookW = 14 + rng() * 22
        const bookH = _shelfHeight * (0.55 + rng() * 0.38)
        const palette = theme.bookPalettes[Math.floor(rng() * theme.bookPalettes.length)]

        if (x + bookW > maxX - 8) break

        books.push({
            x,
            width: bookW,
            height: bookH,
            colors: palette,
            hasTitle: rng() > 0.3,
            isLying: false,
        })

        x += bookW + (rng() < 0.3 ? 1 : 0.5)
    }

    return books
}

function renderBook(
    book: BookSpec,
    shelfY: number,
    _shelfHeight: number,
    theme: BookcaseTheme,
    _rng: () => number,
    id: string
): string {
    if (book.isLying) {
        const bookY = shelfY - book.height - 2
        const col = book.colors[0]
        return `
      <rect x="${book.x}" y="${bookY}" width="${book.width}" height="${book.height}"
        fill="${col}" rx="1"/>
      <rect x="${book.x}" y="${bookY}" width="${book.width}" height="2"
        fill="${lerp(col, '#ffffff', 0.3)}" rx="1"/>
      <rect x="${book.x}" y="${bookY + book.height - 2}" width="${book.width}" height="2"
        fill="${lerp(col, '#000000', 0.4)}" rx="1"/>
    `
    }

    const bookY = shelfY - book.height
    const [baseColor, darkColor, lightColor] = book.colors

    // Spine gradient for depth
    const gradId = `bg-${id}`
    let spineDetails = ''

    if (theme.spineStyle === 'leather') {
        spineDetails = `
      <rect x="${book.x + 2}" y="${bookY + 4}" width="${book.width - 4}" height="1.5"
        fill="${lerp(baseColor, '#ffffff', 0.15)}" opacity="0.6"/>
      <rect x="${book.x + 2}" y="${bookY + book.height - 6}" width="${book.width - 4}" height="1.5"
        fill="${lerp(baseColor, '#ffffff', 0.15)}" opacity="0.6"/>
      <rect x="${book.x + 2}" y="${bookY + book.height * 0.4}" width="${book.width - 4}" height="1"
        fill="${lerp(baseColor, '#000000', 0.2)}" opacity="0.5"/>
    `
    } else if (theme.spineStyle === 'metal') {
        spineDetails = `
      <rect x="${book.x}" y="${bookY}" width="${book.width}" height="${book.height}"
        fill="url(#metalSheen-${id})" opacity="0.4"/>
    `
    } else if (theme.spineStyle === 'circuit') {
        const cx = book.x + book.width / 2
        spineDetails = `
      <line x1="${cx}" y1="${bookY + 8}" x2="${cx}" y2="${bookY + book.height - 8}"
        stroke="${lerp(baseColor, '#00ffff', 0.8)}" stroke-width="0.5" opacity="0.6"/>
      <circle cx="${cx}" cy="${bookY + book.height * 0.3}" r="1.5"
        fill="${lerp(baseColor, '#00ffff', 0.9)}" opacity="0.7"/>
      <circle cx="${cx}" cy="${bookY + book.height * 0.7}" r="1.5"
        fill="${lerp(baseColor, '#00ffff', 0.9)}" opacity="0.7"/>
    `
    } else if (theme.spineStyle === 'cracked') {
        spineDetails = `
      <line x1="${book.x + book.width * 0.3}" y1="${bookY + book.height * 0.2}"
        x2="${book.x + book.width * 0.7}" y2="${bookY + book.height * 0.5}"
        stroke="${lerp(baseColor, '#000000', 0.5)}" stroke-width="0.5" opacity="0.7"/>
      <line x1="${book.x + book.width * 0.6}" y1="${bookY + book.height * 0.4}"
        x2="${book.x + book.width * 0.4}" y2="${bookY + book.height * 0.8}"
        stroke="${lerp(baseColor, '#000000', 0.5)}" stroke-width="0.5" opacity="0.5"/>
    `
    } else if (theme.spineStyle === 'glossy') {
        spineDetails = `
      <rect x="${book.x}" y="${bookY}" width="${book.width * 0.35}" height="${book.height}"
        fill="${lerp(baseColor, '#ffffff', 0.15)}" rx="1" opacity="0.3"/>
    `
    }

    // Title line
    const titleLine = book.hasTitle ? `
    <rect x="${book.x + 3}" y="${bookY + book.height * 0.35}" width="${book.width - 6}" height="1.5"
      fill="${lerp(baseColor, '#ffffff', 0.5)}" opacity="0.7" rx="0.5"/>
    <rect x="${book.x + 3}" y="${bookY + book.height * 0.42}" width="${book.width - 6}" height="1"
      fill="${lerp(baseColor, '#ffffff', 0.4)}" opacity="0.5" rx="0.5"/>
  ` : ''

    // Shadow on right edge
    const shadowLine = `
    <rect x="${book.x + book.width - 2}" y="${bookY}" width="2" height="${book.height}"
      fill="${lerp(darkColor, '#000000', 0.3)}" opacity="0.6"/>
  `

    // Highlight on left edge
    const highlightLine = `
    <rect x="${book.x}" y="${bookY}" width="1.5" height="${book.height}"
      fill="${lerp(lightColor, '#ffffff', 0.2)}" opacity="0.4"/>
  `

    // Top of book (pages)
    const pageTop = `
    <rect x="${book.x + 1}" y="${bookY}" width="${book.width - 2}" height="2"
      fill="${lerp(baseColor, '#f0e8d8', 0.6)}" opacity="0.5"/>
  `

    return `
    <defs>
      <linearGradient id="${gradId}" x1="0" y1="0" x2="1" y2="0">
        <stop offset="0%" stop-color="${lightColor}" stop-opacity="0.9"/>
        <stop offset="30%" stop-color="${baseColor}" stop-opacity="1"/>
        <stop offset="100%" stop-color="${darkColor}" stop-opacity="1"/>
      </linearGradient>
      <linearGradient id="metalSheen-${id}" x1="0" y1="0" x2="1" y2="0">
        <stop offset="0%" stop-color="#ffffff" stop-opacity="0"/>
        <stop offset="40%" stop-color="#ffffff" stop-opacity="0.15"/>
        <stop offset="60%" stop-color="#ffffff" stop-opacity="0.05"/>
        <stop offset="100%" stop-color="#ffffff" stop-opacity="0"/>
      </linearGradient>
    </defs>
    <rect x="${book.x}" y="${bookY}" width="${book.width}" height="${book.height}"
      fill="url(#${gradId})" rx="1"/>
    ${spineDetails}
    ${titleLine}
    ${shadowLine}
    ${highlightLine}
    ${pageTop}
  `
}

function generateBookcaseSVG(themeId: ThemeId, width: number, height: number): string {
    const theme = bookcaseThemes[themeId]
    const rng = seededRandom(themeId.split('').reduce((a, c) => a + c.charCodeAt(0), 0))

    const frameW = 16 // side frame width
    const shelfCount = Math.ceil(height / 160)
    const shelfH = height / shelfCount
    const shelfThickness = 14
    const topFrameH = 20
    const bottomFrameH = 24

    // Wood grain filter for wooden themes
    const woodGrainFilter = ['woody', 'dark-academia', 'post-apocalyptic'].includes(themeId) ? `
    <filter id="woodGrain" x="0" y="0" width="100%" height="100%">
      <feTurbulence type="fractalNoise" baseFrequency="0.015 0.8" numOctaves="4" seed="42" result="noise"/>
      <feColorMatrix type="saturate" values="0" in="noise" result="grayNoise"/>
      <feBlend in="SourceGraphic" in2="grayNoise" mode="multiply" result="blend"/>
      <feComponentTransfer in="blend">
        <feFuncA type="linear" slope="1"/>
      </feComponentTransfer>
    </filter>
  ` : ''

    const metalFilter = themeId === 'metallic' ? `
    <filter id="metalFilter">
      <feTurbulence type="fractalNoise" baseFrequency="0.65 0.02" numOctaves="3" seed="5" result="noise"/>
      <feColorMatrix type="saturate" values="0" in="noise" result="grayNoise"/>
      <feBlend in="SourceGraphic" in2="grayNoise" mode="overlay" result="blend"/>
    </filter>
  ` : ''

    const noiseFilter = `
    <filter id="grainFilter">
      <feTurbulence type="fractalNoise" baseFrequency="0.85" numOctaves="4" stitchTiles="stitch" result="noise"/>
      <feColorMatrix type="saturate" values="0" in="noise" result="gray"/>
      <feBlend in="SourceGraphic" in2="gray" mode="multiply" result="blended"/>
      <feComponentTransfer in="blended">
        <feFuncA type="linear" slope="${1 - theme.grainOpacity}"/>
      </feComponentTransfer>
    </filter>
  `

    // Outer vignette gradient
    const vignetteGrad = `
    <radialGradient id="vignette" cx="50%" cy="50%" r="70%">
      <stop offset="0%" stop-color="transparent"/>
      <stop offset="100%" stop-color="#000000" stop-opacity="${0.3 + theme.darkness * 0.5}"/>
    </radialGradient>
  `

    // Shelf wood grain gradient
    const shelfGrad = `
    <linearGradient id="shelfGrad" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0%" stop-color="${theme.shelfHighlight}"/>
      <stop offset="30%" stop-color="${theme.shelfColor}"/>
      <stop offset="100%" stop-color="${theme.shelfShadow}"/>
    </linearGradient>
    <linearGradient id="frameGrad" x1="0" y1="0" x2="1" y2="0">
      <stop offset="0%" stop-color="${theme.frameHighlight}"/>
      <stop offset="40%" stop-color="${theme.frameColor}"/>
      <stop offset="100%" stop-color="${theme.frameShadow}"/>
    </linearGradient>
    <linearGradient id="frameGradR" x1="0" y1="0" x2="1" y2="0">
      <stop offset="0%" stop-color="${theme.frameShadow}"/>
      <stop offset="60%" stop-color="${theme.frameColor}"/>
      <stop offset="100%" stop-color="${theme.frameHighlight}"/>
    </linearGradient>
  `

    // Ambient glow for themed lighting
    const ambientGlow = `
    <radialGradient id="ambientGlow" cx="50%" cy="30%" r="60%">
      <stop offset="0%" stop-color="${theme.ambientColor}" stop-opacity="${theme.ambientOpacity}"/>
      <stop offset="100%" stop-color="${theme.ambientColor}" stop-opacity="0"/>
    </radialGradient>
  `

    // Shelf edge shadow gradient
    const shelfEdgeShadow = `
    <linearGradient id="shelfEdgeShadow" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0%" stop-color="#000000" stop-opacity="0.5"/>
      <stop offset="40%" stop-color="#000000" stop-opacity="0.15"/>
      <stop offset="100%" stop-color="#000000" stop-opacity="0"/>
    </linearGradient>
  `

    // Futuristic scanline pattern
    const scanlines = themeId === 'futuristic' ? `
    <pattern id="scanlines" x="0" y="0" width="1" height="4" patternUnits="userSpaceOnUse">
      <rect width="1" height="2" fill="#000000" opacity="0.15"/>
    </pattern>
  ` : ''

    // Generate all shelves and their books
    let shelvesContent = ''
    let booksContent = ''

    for (let s = 0; s < shelfCount; s++) {
        const shelfY = topFrameH + (s + 1) * shelfH - shelfThickness
        const availH = shelfH - shelfThickness - 8
        const innerW = width - frameW * 2

        // Back wall section
        shelvesContent += `
      <rect x="${frameW}" y="${topFrameH + s * shelfH}" width="${innerW}" height="${shelfH}"
        fill="${theme.wallColor}"/>
    `

        // Wall texture lines for wood themes
        if (['woody', 'dark-academia'].includes(themeId)) {
            for (let i = 0; i < 8; i++) {
                const lx = frameW + rng() * innerW
                shelvesContent += `
          <line x1="${lx}" y1="${topFrameH + s * shelfH}" x2="${lx + rng() * 20 - 10}"
            y2="${topFrameH + (s + 1) * shelfH}"
            stroke="${theme.shelfHighlight}" stroke-width="${0.3 + rng() * 0.4}" opacity="${0.08 + rng() * 0.06}"/>
        `
            }
        }

        // Circuit lines for futuristic
        if (themeId === 'futuristic') {
            const circY = topFrameH + s * shelfH + availH * 0.5
            shelvesContent += `
        <line x1="${frameW + 20}" y1="${circY}" x2="${width - frameW - 20}" y2="${circY}"
          stroke="#00f0ff" stroke-width="0.5" opacity="0.08"/>
        <circle cx="${frameW + 30}" cy="${circY}" r="2" fill="#00f0ff" opacity="0.12"/>
        <circle cx="${width - frameW - 30}" cy="${circY}" r="2" fill="#00f0ff" opacity="0.12"/>
      `
        }

        // Stars on back wall for space theme
        if (themeId === 'space') {
            for (let i = 0; i < 12; i++) {
                const sx = frameW + rng() * innerW
                const sy = topFrameH + s * shelfH + rng() * (shelfH - shelfThickness)
                const sr = 0.3 + rng() * 0.8
                shelvesContent += `<circle cx="${sx}" cy="${sy}" r="${sr}" fill="white" opacity="${0.3 + rng() * 0.5}"/>`
            }
        }

        // Generate and render books for this shelf
        const shelfBooks = generateShelf(rng, innerW, availH, theme, 0.06)
        shelfBooks.forEach((book, bi) => {
            const adjustedBook = { ...book, x: book.x + frameW }
            booksContent += renderBook(adjustedBook, shelfY, availH, theme, rng, `s${s}b${bi}`)
        })

        // Shelf surface
        shelvesContent += `
      <rect x="${frameW - 2}" y="${shelfY}" width="${innerW + 4}" height="${shelfThickness}"
        fill="url(#shelfGrad)"/>
    `

        // Shelf front edge highlight
        shelvesContent += `
      <rect x="${frameW - 2}" y="${shelfY}" width="${innerW + 4}" height="2"
        fill="${theme.shelfHighlight}" opacity="0.7"/>
    `

        // Shelf front edge dark line
        shelvesContent += `
      <rect x="${frameW - 2}" y="${shelfY + shelfThickness - 2}" width="${innerW + 4}" height="2"
        fill="${theme.shelfShadow}" opacity="0.8"/>
    `

        // Shelf edge border
        shelvesContent += `
      <rect x="${frameW - 2}" y="${shelfY}" width="${innerW + 4}" height="${shelfThickness}"
        fill="none" stroke="${theme.shelfEdge}" stroke-width="0.5" opacity="0.5"/>
    `

        // Drop shadow below shelf onto books
        shelvesContent += `
      <rect x="${frameW}" y="${shelfY + shelfThickness}" width="${innerW}" height="20"
        fill="url(#shelfEdgeShadow)"/>
    `

        // Shelf wood grain texture lines
        if (['woody', 'dark-academia', 'post-apocalyptic'].includes(themeId)) {
            for (let i = 0; i < 6; i++) {
                const ly = shelfY + 3 + i * 2
                shelvesContent += `
          <line x1="${frameW}" y1="${ly}" x2="${width - frameW}" y2="${ly + rng() * 2 - 1}"
            stroke="${theme.shelfHighlight}" stroke-width="0.4" opacity="${0.15 + rng() * 0.1}"/>
        `
            }
        }

        // Metal shelf lines
        if (themeId === 'metallic') {
            shelvesContent += `
        <rect x="${frameW}" y="${shelfY + 4}" width="${innerW}" height="1"
          fill="#ffffff" opacity="0.08"/>
        <rect x="${frameW}" y="${shelfY + 6}" width="${innerW}" height="0.5"
          fill="#ffffff" opacity="0.04"/>
      `
            // Bolt details
            for (let b = 0; b < 6; b++) {
                const bx = frameW + 20 + b * (innerW - 40) / 5
                shelvesContent += `
          <circle cx="${bx}" cy="${shelfY + shelfThickness / 2}" r="2.5"
            fill="${theme.shelfColor}" stroke="${theme.shelfHighlight}" stroke-width="0.8" opacity="0.8"/>
          <circle cx="${bx}" cy="${shelfY + shelfThickness / 2}" r="1"
            fill="${theme.shelfShadow}" opacity="0.6"/>
        `
            }
        }

        // Futuristic shelf glow
        if (themeId === 'futuristic') {
            shelvesContent += `
        <rect x="${frameW}" y="${shelfY}" width="${innerW}" height="2"
          fill="#00f0ff" opacity="0.3"/>
        <rect x="${frameW}" y="${shelfY + shelfThickness - 1}" width="${innerW}" height="1"
          fill="#00f0ff" opacity="0.2"/>
      `
        }

        // Space shelf glow
        if (themeId === 'space') {
            shelvesContent += `
        <rect x="${frameW}" y="${shelfY}" width="${innerW}" height="1.5"
          fill="#a060ff" opacity="0.4"/>
      `
        }
    }

    // Left and right frame
    const frames = `
    <rect x="0" y="0" width="${frameW}" height="${height}" fill="url(#frameGrad)"/>
    <rect x="${width - frameW}" y="0" width="${frameW}" height="${height}" fill="url(#frameGradR)"/>
    <!-- Frame inner shadow -->
    <rect x="${frameW}" y="0" width="8" height="${height}"
      fill="url(#frameGrad)" opacity="0.3"/>
    <rect x="${width - frameW - 8}" y="0" width="8" height="${height}"
      fill="url(#frameGradR)" opacity="0.3"/>
    <!-- Frame highlight edge -->
    <rect x="0" y="0" width="2" height="${height}"
      fill="${theme.frameHighlight}" opacity="0.6"/>
    <rect x="${width - 2}" y="0" width="2" height="${height}"
      fill="${theme.frameHighlight}" opacity="0.6"/>
    <!-- Frame outer shadow -->
    <rect x="${frameW - 2}" y="0" width="3" height="${height}"
      fill="${theme.frameShadow}" opacity="0.5"/>
    <rect x="${width - frameW - 1}" y="0" width="3" height="${height}"
      fill="${theme.frameShadow}" opacity="0.5"/>
  `

    // Top and bottom frame
    const topBottom = `
    <rect x="0" y="0" width="${width}" height="${topFrameH}" fill="url(#shelfGrad)"/>
    <rect x="0" y="${height - bottomFrameH}" width="${width}" height="${bottomFrameH}" fill="url(#shelfGrad)"/>
    <!-- Top frame details -->
    <rect x="0" y="0" width="${width}" height="2" fill="${theme.frameHighlight}" opacity="0.8"/>
    <rect x="0" y="${topFrameH - 2}" width="${width}" height="2" fill="${theme.shelfShadow}" opacity="0.6"/>
    <!-- Bottom frame details -->
    <rect x="0" y="${height - bottomFrameH}" width="${width}" height="2" fill="${theme.shelfHighlight}" opacity="0.6"/>
    <rect x="0" y="${height - 2}" width="${width}" height="2" fill="${theme.shelfShadow}" opacity="0.8"/>
    <!-- Baseboard bumps for wood themes -->
  `

    // Frame ornamental details per theme
    let frameOrnaments = ''
    if (['woody', 'dark-academia'].includes(themeId)) {
        // Wood frame decorative grooves
        frameOrnaments += `
      <rect x="3" y="${topFrameH + 10}" width="10" height="${height - topFrameH - bottomFrameH - 20}"
        fill="none" stroke="${theme.shelfHighlight}" stroke-width="0.5" opacity="0.3" rx="2"/>
      <rect x="${width - 13}" y="${topFrameH + 10}" width="10" height="${height - topFrameH - bottomFrameH - 20}"
        fill="none" stroke="${theme.shelfHighlight}" stroke-width="0.5" opacity="0.3" rx="2"/>
    `
    } else if (themeId === 'metallic') {
        // Metal frame rivets
        for (let r = 0; r < shelfCount + 1; r++) {
            const ry = topFrameH + r * shelfH - 8
            frameOrnaments += `
        <circle cx="8" cy="${ry}" r="3" fill="${theme.shelfColor}"
          stroke="${theme.shelfHighlight}" stroke-width="0.8" opacity="0.9"/>
        <circle cx="8" cy="${ry}" r="1.2" fill="${theme.shelfShadow}" opacity="0.7"/>
        <circle cx="${width - 8}" cy="${ry}" r="3" fill="${theme.shelfColor}"
          stroke="${theme.shelfHighlight}" stroke-width="0.8" opacity="0.9"/>
        <circle cx="${width - 8}" cy="${ry}" r="1.2" fill="${theme.shelfShadow}" opacity="0.7"/>
      `
        }
    } else if (themeId === 'futuristic') {
        // LED strips on frame
        frameOrnaments += `
      <rect x="5" y="${topFrameH}" width="4" height="${height - topFrameH - bottomFrameH}"
        fill="#00f0ff" opacity="0.15" rx="2"/>
      <rect x="${width - 9}" y="${topFrameH}" width="4" height="${height - topFrameH - bottomFrameH}"
        fill="#00f0ff" opacity="0.15" rx="2"/>
    `
        for (let r = 0; r < shelfCount; r++) {
            const ry = topFrameH + r * shelfH + shelfH / 2
            frameOrnaments += `
        <circle cx="7" cy="${ry}" r="2" fill="#00f0ff" opacity="0.6"/>
        <circle cx="${width - 7}" cy="${ry}" r="2" fill="#00f0ff" opacity="0.6"/>
      `
        }
    } else if (themeId === 'space') {
        frameOrnaments += `
      <rect x="5" y="${topFrameH}" width="3" height="${height - topFrameH - bottomFrameH}"
        fill="#a060ff" opacity="0.2" rx="2"/>
      <rect x="${width - 8}" y="${topFrameH}" width="3" height="${height - topFrameH - bottomFrameH}"
        fill="#a060ff" opacity="0.2" rx="2"/>
    `
    }

    // Scanlines overlay for futuristic
    const scanlineOverlay = themeId === 'futuristic' ? `
    <rect x="0" y="0" width="${width}" height="${height}" fill="url(#scanlines)" opacity="0.4"/>
  ` : ''

    // Ambient overlay
    const ambientOverlay = `
    <rect x="0" y="0" width="${width}" height="${height}" fill="url(#ambientGlow)"/>
  `

    // Final vignette
    const vignetteOverlay = `
    <rect x="0" y="0" width="${width}" height="${height}" fill="url(#vignette)"/>
  `

    return `<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}">
  <defs>
    ${woodGrainFilter}
    ${metalFilter}
    ${noiseFilter}
    ${shelfGrad}
    ${vignetteGrad}
    ${ambientGlow}
    ${shelfEdgeShadow}
    ${scanlines}
  </defs>

  <!-- Background fill -->
  <rect width="${width}" height="${height}" fill="${theme.wallColor}"/>

  <!-- Shelves and wall sections -->
  ${shelvesContent}

  <!-- Books -->
  ${booksContent}

  <!-- Frame sides -->
  ${frames}

  <!-- Top and bottom frame -->
  ${topBottom}

  <!-- Frame ornaments -->
  ${frameOrnaments}

  <!-- Scanlines -->
  ${scanlineOverlay}

  <!-- Ambient light -->
  ${ambientOverlay}

  <!-- Vignette -->
  ${vignetteOverlay}
</svg>`
}

// ─── Component ───────────────────────────────────────────────────────────────

export default function ThemeBackground() {
    const { themeId } = useTheme()

    const svgContent = useMemo(() => {
        if (themeId === 'post-apocalyptic' || themeId === 'woody' ) return ''   // skip SVG generation
        const w = window.innerWidth || 1920
        const h = window.innerHeight || 1080
        return generateBookcaseSVG(themeId, w, h)
    }, [themeId])

    const svgUrl = useMemo(() => {
        if (themeId === 'post-apocalyptic' || themeId === 'woody') return ''
        const blob = new Blob([svgContent], { type: 'image/svg+xml' })
        return URL.createObjectURL(blob)
    }, [svgContent])

    if (themeId === 'post-apocalyptic') {
        return (
            <div style={{
                position: 'fixed', top: 0, left: 0,
                width: '100vw', height: '100vh',
                zIndex: -10, pointerEvents: 'none',
            }}>
                {/* ── 1. Photo base ─────────────────────────────────── */}
                <div style={{
                    position: 'absolute', inset: 0,
                    backgroundImage: 'url("/apocalyptic_library.png")',
                    backgroundSize: 'cover',
                    backgroundPosition: 'center top',
                }} />

                {/* ── 2. Deep vignette — kills the bright edges ──────── */}
                <div style={{
                    position: 'absolute', inset: 0,
                    background: [
                        'radial-gradient(ellipse 80% 70% at 58% 42%, rgba(12,9,6,0.15) 0%, rgba(12,9,6,0.65) 60%, rgba(12,9,6,0.92) 100%)',
                    ].join(', '),
                }} />

                {/* ── 3. Top + bottom crush — keeps UI legible ──────── */}
                <div style={{
                    position: 'absolute', inset: 0,
                    background: 'linear-gradient(to bottom, rgba(12,9,6,0.85) 0%, transparent 18%, transparent 78%, rgba(12,9,6,0.90) 100%)',
                }} />

                {/* ── 4. Left edge darkening for sidebar contrast ────── */}
                <div style={{
                    position: 'absolute', inset: 0,
                    background: 'linear-gradient(to right, rgba(12,9,6,0.80) 0%, rgba(12,9,6,0.30) 220px, transparent 400px)',
                }} />

                {/* ── 5. Animated dust motes (pure CSS, no JS) ─────── */}
                <svg
                    style={{ position: 'absolute', inset: 0, width: '100%', height: '100%', opacity: 0.55 }}
                    xmlns="http://www.w3.org/2000/svg"
                >
                    <defs>
                        <radialGradient id="mote" cx="50%" cy="50%" r="50%">
                            <stop offset="0%" stopColor="#ece3d4" stopOpacity="1" />
                            <stop offset="100%" stopColor="#ece3d4" stopOpacity="0" />
                        </radialGradient>
                        <style>{`
                            @keyframes drift1 { 0%{transform:translate(0,0) scale(1)} 50%{transform:translate(18px,-28px) scale(1.3)} 100%{transform:translate(0,0) scale(1)} }
                            @keyframes drift2 { 0%{transform:translate(0,0) scale(0.8)} 50%{transform:translate(-14px,20px) scale(1.1)} 100%{transform:translate(0,0) scale(0.8)} }
                            @keyframes drift3 { 0%{transform:translate(0,0)} 33%{transform:translate(10px,-16px)} 66%{transform:translate(-8px,10px)} 100%{transform:translate(0,0)} }
                            @keyframes flicker { 0%,100%{opacity:0.6} 50%{opacity:1} }
                            .m1{animation:drift1 9s ease-in-out infinite}
                            .m2{animation:drift2 13s ease-in-out infinite}
                            .m3{animation:drift3 7s ease-in-out infinite}
                            .m4{animation:drift1 11s ease-in-out infinite reverse}
                            .m5{animation:drift2 8s ease-in-out infinite reverse}
                            .mf{animation:flicker 3s ease-in-out infinite}
                        `}</style>
                    </defs>
                    {/* Motes positioned near the light beam area (55–65% x) */}
                    <circle className="m1" cx="58%" cy="38%" r="2.5" fill="url(#mote)" />
                    <circle className="m2" cx="61%" cy="44%" r="1.8" fill="url(#mote)" />
                    <circle className="m3" cx="56%" cy="52%" r="3"   fill="url(#mote)" />
                    <circle className="m4" cx="63%" cy="35%" r="1.5" fill="url(#mote)" />
                    <circle className="m5" cx="59%" cy="60%" r="2"   fill="url(#mote)" />
                    <circle className="m1 mf" cx="57%" cy="48%" r="1.2" fill="url(#mote)" />
                    <circle className="m3" cx="64%" cy="55%" r="2.2" fill="url(#mote)" />
                    <circle className="m2" cx="55%" cy="40%" r="1.6" fill="url(#mote)" />
                    {/* A few more scattered */}
                    <circle className="m4" cx="42%" cy="30%" r="1.2" fill="url(#mote)" />
                    <circle className="m5" cx="72%" cy="42%" r="1.4" fill="url(#mote)" />
                    <circle className="m1" cx="38%" cy="55%" r="1"   fill="url(#mote)" />
                    <circle className="m2" cx="78%" cy="58%" r="1.3" fill="url(#mote)" />
                </svg>

                {/* ── 6. Film grain overlay ─────────────────────────── */}
                <div style={{
                    position: 'absolute', inset: 0,
                    backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='200' height='200'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.75' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='200' height='200' filter='url(%23n)' opacity='0.08'/%3E%3C/svg%3E")`,
                    backgroundRepeat: 'repeat',
                    backgroundSize: '200px',
                    mixBlendMode: 'overlay',
                    opacity: 0.6,
                }} />
            </div>
        )
    }

    if (themeId === 'woody') {
        return (
            <div style={{
                position: 'fixed', top: 0, left: 0,
                width: '100vw', height: '100vh',
                zIndex: -10, pointerEvents: 'none',
            }}>
                {/* ── 1. Photo base ─────────────────────────────────────── */}
                <div style={{
                    position: 'absolute', inset: 0,
                    backgroundImage: 'url("/woody_library.png")',
                    backgroundSize: 'cover',
                    backgroundPosition: 'center top',
                }} />

                {/* ── 2. Radial vignette — darkens edges, preserves centre glow */}
                <div style={{
                    position: 'absolute', inset: 0,
                    background: 'radial-gradient(ellipse 75% 65% at 50% 30%, rgba(14,9,6,0.10) 0%, rgba(14,9,6,0.55) 55%, rgba(14,9,6,0.92) 100%)',
                }} />

                {/* ── 3. Top crush — ceiling glow bleeds slightly through */}
                <div style={{
                    position: 'absolute', inset: 0,
                    background: 'linear-gradient(to bottom, rgba(14,9,6,0.50) 0%, transparent 20%, transparent 75%, rgba(14,9,6,0.92) 100%)',
                }} />

                {/* ── 4. Left edge for sidebar contrast ───────────────── */}
                <div style={{
                    position: 'absolute', inset: 0,
                    background: 'linear-gradient(to right, rgba(14,9,6,0.82) 0%, rgba(14,9,6,0.35) 220px, transparent 420px)',
                }} />

                {/* ── 5. Fireplace flicker: warm amber pulse on the glow ─ */}
                <div style={{
                    position: 'absolute', inset: 0,
                    background: 'radial-gradient(ellipse 40% 30% at 50% 0%, rgba(196,128,58,0.12) 0%, transparent 100%)',
                    animation: 'woodyFlicker 4s ease-in-out infinite',
                }} />

                {/* ── 6. Fine film grain ───────────────────────────────── */}
                <div style={{
                    position: 'absolute', inset: 0,
                    backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='180' height='180'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.72' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='180' height='180' filter='url(%23n)' opacity='0.07'/%3E%3C/svg%3E")`,
                    backgroundRepeat: 'repeat',
                    backgroundSize: '180px',
                    mixBlendMode: 'overlay',
                    opacity: 0.5,
                }} />

                {/* ── 7. Keyframe style tag ────────────────────────────── */}
                <style>{`
                    @keyframes woodyFlicker {
                        0%,100% { opacity: 0.7; }
                        25%     { opacity: 1.0; }
                        50%     { opacity: 0.8; }
                        75%     { opacity: 0.95; }
                    }
                `}</style>
            </div>
        )
    }

    return (
        <div style={{
            position: 'fixed', top: 0, left: 0,
            width: '100vw', height: '100vh',
            zIndex: -10, pointerEvents: 'none',
            backgroundImage: `url("${svgUrl}")`,
            backgroundSize: 'cover',
            backgroundPosition: 'center',
        }} />
    )
}