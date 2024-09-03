/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [    "./client/ui/templates/**/*.{html,gohtml,tmpl,svg,js}",
    "./client/ui/assets/js/**/*.js",
    "./client/ui/**/*.{html,gohtml,tmpl,svg,js}"], // This line covers all UI files],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        primary: '#f6ad55',
      }
    }
  },
  plugins: [],
}
