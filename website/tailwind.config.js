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
        primary: {
          DEFAULT: 'rgb(var(--color-primary-rgb) / <alpha-value>)',
          dark: 'rgb(var(--color-primary-dark-rgb) / <alpha-value>)',
        },
        surface: {
          bg: '#f7f6f3',
          card: '#ffffff',
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
