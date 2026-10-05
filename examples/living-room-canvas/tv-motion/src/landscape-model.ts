export type Setting = readonly [number, number];

// A toy model predicts (u², v - 0.35u, u), with target (1, 0, 1).
// The weighted squared prediction error creates two valleys.
export const loss = ([u, v]: Setting) =>
  (u * u - 1) ** 2 + 0.6 * (v - 0.35 * u) ** 2 + 0.15 * (u - 1) ** 2;

export const gradient = ([u, v]: Setting): Setting => [
  4 * u * (u * u - 1) - 0.42 * (v - 0.35 * u) + 0.3 * (u - 1),
  1.2 * (v - 0.35 * u),
];

export const trajectory: Setting[] = [[1.6, 1.1]];
for (let i = 0; i < 40; i++) {
  const previous = trajectory[trajectory.length - 1];
  const g = gradient(previous);
  trajectory.push([previous[0] - 0.04 * g[0], previous[1] - 0.04 * g[1]]);
}

export const height = (p: Setting) => loss(p) * 0.18;
export const clamp = (n: number, low: number, high: number) => Math.min(high, Math.max(low, n));
export const mix = (a: number, b: number, t: number) => a + (b - a) * t;
export const smooth = (t: number) => t * t * (3 - 2 * t);

export function progress(frame: number) {
  if (frame < 1080) return 0;
  if (frame < 1440) return smooth(clamp((frame - 1095) / 90, 0, 1));
  // One deliberate two-second update at a time. Holds separate the steps.
  const local = clamp(frame - 1455, 0, 630);
  const step = Math.floor(local / 90);
  return 1 + step + smooth(clamp((local % 90) / 60, 0, 1));
}

export function settingAt(amount: number): Setting {
  const index = Math.floor(amount), t = amount - index;
  const a = trajectory[index], b = trajectory[Math.min(index + 1, trajectory.length - 1)];
  return [mix(a[0], b[0], t), mix(a[1], b[1], t)];
}

export function cameraAt(frame: number) {
  const t = smooth(clamp((frame - 720) / 150, 0, 1));
  return {
    position: [mix(3.8, 3.5, t), mix(8.0, 7.6, t), mix(4.2, 3.8, t)] as [number, number, number],
    target: [mix(0, 0.2, t), 0.25, mix(0, 0.15, t)] as [number, number, number],
  };
}
