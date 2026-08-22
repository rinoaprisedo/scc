/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{js,jsx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['DM Sans', 'sans-serif'],
        mono: ['DM Mono', 'monospace'],
      },
      colors: {
        navy: {
          DEFAULT: '#0b1830',
          light: '#142a4d',
          dark: '#060d1b',
        },
        gold: {
          DEFAULT: '#d4af5a',
          light: '#f0d385',
          dark: '#a97f2e',
        },
        surface: {
          bg: '#f7f6f3',
          card: '#ffffff',
          border: '#e2e1db',
        },
        text: {
          primary: '#1a1917',
          secondary: '#6b6a65',
        },
      },
    },
  },
  plugins: [],
}
