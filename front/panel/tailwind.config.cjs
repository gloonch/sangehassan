module.exports = {
  content: ["./index.html", "./src/**/*.{js,jsx}", "../shared/**/*.{js,jsx,json}"],
  theme: {
    extend: {
      colors: {
        primary: "#083A4F",
        accent: "#A58D66",
        sand: "#E5E1DD"
      },
      fontFamily: {
        display: ["Playfair Display", "Noto Sans Arabic", "serif"],
        body: ["Manrope", "Noto Sans Arabic", "sans-serif"]
      }
    }
  },
  plugins: []
};
