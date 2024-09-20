/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [    "./client/ui/templates/**/*.{html,gohtml,tmpl,svg,js}",
    "./client/ui/assets/js/**/*.js",
    "./client/ui/**/*.{html,gohtml,tmpl,svg,js}"], // This line covers all UI files],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        red: {
          primary: '#f6ad55',
          600: '#dc2626', // Custom red color
        },
      },
      animation: {
        line: 'lineAnimation 5s linear forwards',
      },
      keyframes: {
        lineAnimation: {
          '0%': { width: '100%' },
          '100%': { width: '0%' },
        },
      },
    },
  },
  plugins: [],
}
