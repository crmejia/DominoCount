/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./templates/*.html"],
  theme: {
    extend: {
      colors: {
        firstcolor: '#feffdf',
        secondcolor: '#dde0ab',
        thirdcolor: '#97cba9',
        fourthcolor: '#668ba4',
      }
    },
    plugins: [],
  }
}