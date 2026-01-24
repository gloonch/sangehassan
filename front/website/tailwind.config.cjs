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
      },
      keyframes: {
        marquee: {
          "0%": { transform: "translateX(0%)" },
          "100%": { transform: "translateX(-50%)" }
        },
        floatIn: {
          "0%": { opacity: "0", transform: "translateY(20px)" },
          "100%": { opacity: "1", transform: "translateY(0)" }
        }
      },
      animation: {
        marquee: "marquee 18s linear infinite",
        floatIn: "floatIn 0.8s ease-out forwards"
      }
    }
  },
  plugins: []
};
