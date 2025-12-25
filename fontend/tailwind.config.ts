
import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      colors: {
        primary: "#7C3AED",
        secondary: "#22D3EE",
        accent: "#0EA5E9",
      },
    },
  },
  plugins: [],
};

export default config;