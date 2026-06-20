/** @type {import('tailwindcss').Config} */

export default {
  darkMode: "class",
  content: ["./index.html", "./src/**/*.{js,ts,vue}"],
  theme: {
    container: {
      center: true,
    },
    extend: {
      colors: {
        ink: {
          950: "#070E1C",
          900: "#0A1222",
          800: "#0E1A2E",
          700: "#111E33",
          600: "#16263F",
          500: "#1E3050",
        },
        pv: {
          DEFAULT: "#2EE6A6",
          dim: "#1FB580",
          deep: "#0E5C45",
        },
        amber: { glow: "#FFB020" },
        danger: { glow: "#FF4D4F" },
        cyan: { glow: "#38BDF8" },
        fog: { DEFAULT: "#8AA0BD", dim: "#5E7299" },
      },
      fontFamily: {
        display: ['"Chakra Petch"', "sans-serif"],
        mono: ['"IBM Plex Mono"', "monospace"],
        sans: ['"Noto Sans SC"', "sans-serif"],
      },
      boxShadow: {
        glow: "0 0 0 1px rgba(46,230,166,.35), 0 0 24px -6px rgba(46,230,166,.45)",
        panel: "0 1px 0 0 rgba(255,255,255,.02), 0 12px 40px -20px rgba(0,0,0,.8)",
        amber: "0 0 0 1px rgba(255,176,32,.4), 0 0 24px -6px rgba(255,176,32,.4)",
      },
      backgroundImage: {
        grid: "linear-gradient(rgba(46,230,166,.05) 1px, transparent 1px), linear-gradient(90deg, rgba(46,230,166,.05) 1px, transparent 1px)",
      },
      keyframes: {
        pulseAlarm: {
          "0%,100%": { opacity: "1" },
          "50%": { opacity: ".35" },
        },
        sweep: {
          "0%": { transform: "translateX(-100%)" },
          "100%": { transform: "translateX(100%)" },
        },
      },
      animation: {
        pulseAlarm: "pulseAlarm 1.6s ease-in-out infinite",
        sweep: "sweep 2.4s linear infinite",
      },
    },
  },
  plugins: [],
};
