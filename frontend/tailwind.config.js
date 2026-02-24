/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{svelte,js,ts,jsx,tsx}'
  ],
  theme: {
    extend: {
      colors: {
        'bg-primary': '#11111b',
        'bg-secondary': '#1e1e2e',
        'bg-surface': '#282840',
        'bg-hover': '#313150',
        'text-primary': '#cdd6f4',
        'text-secondary': '#a6adc8',
        'text-muted': '#6c7086',
        'accent': '#89b4fa',
        'accent-hover': '#74c7ec',
        'green': '#a6e3a1',
        'yellow': '#f9e2af',
        'red': '#f38ba8',
        'orange': '#fab387',
        'mauve': '#cba6f7',
        'border': '#45475a',
      },
      borderRadius: {
        'DEFAULT': '8px',
        'lg': '12px',
      },
    },
  },
  plugins: [],
}
