/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        navy: {
          DEFAULT: '#162447',
          light: '#1f3a6e',
        },
      },
    },
  },
  plugins: [],
}

