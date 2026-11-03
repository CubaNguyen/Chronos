/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        dark: {
          bg: '#090d16',
          sidebar: '#0d1322',
          card: '#131b2e',
          border: '#1e293b',
          input: '#0b1120',
        }
      }
    },
  },
  plugins: [],
}

