
const bv2rgb = (bv: number): string => {
  bv = Math.max(-0.4, Math.min(2, bv));
  let t;
  return `#${[
    bv < 0 ? (t = (bv + 0.4) / 0.4, 0.61 + (0.11 * t) + (0.1 * t * t))
      : bv < 0.4 ? (t = bv / 0.4, 0.83 + (0.17 * t))
      : 1,
    bv < 0 ? (t = (bv + 0.4) / 0.4, 0.70 + (0.07 * t) + (0.1 * t * t))
      : bv < 0.4 ? (t = bv / 0.4, 0.87 + (0.11 * t))
      : bv < 1.6 ? (t = (bv - 0.4) / 1.20, 0.98 - (0.16 * t))
      : (t = (bv - 1.6) / 0.4, 0.82 - (0.5 * t * t)), 
    bv < 0.4 ? 1
      : bv < 1.5 ? (t = (bv - 0.4) / 1.1, 1 - (0.47 * t) + (0.1 * t * t))
      : bv < 1.94 ? (t = (bv - 1.5) / 0.44, 0.63 - (0.6 * t * t))
      : 0
  ].map(t => Math.round(t * 255).toString(16).padStart(2, "0")).join("")}`;
}

const temperature = (color: number): number => {
  return 4600 * (1 / (0.92 * color + 1.7) + 1 / (0.92 * color + 0.62));
}

export { bv2rgb, temperature };