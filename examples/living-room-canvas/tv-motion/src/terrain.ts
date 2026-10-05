import { BufferGeometry, Color, Float32BufferAttribute, Vector3 } from 'three';
import { clamp, height, loss, Setting } from './landscape-model';

export function terrain() {
  const geometry = new BufferGeometry();
  const segments = 90, positions: number[] = [], colors: number[] = [], indices: number[] = [];
  const low = new Color('#228c87'), middle = new Color('#719cdd'), high = new Color('#ced7ef');
  for (let i = 0; i <= segments; i++) {
    for (let j = 0; j <= segments; j++) {
      const p: Setting = [-1.85 + 3.7 * i / segments, -1.85 + 3.7 * j / segments];
      positions.push(p[0], height(p), p[1]);
      const t = clamp(loss(p) / 7, 0, 1);
      const color = t < 0.25 ? low.clone().lerp(middle, t * 4) : middle.clone().lerp(high, (t - 0.25) / 0.75);
      colors.push(color.r, color.g, color.b);
      if (i < segments && j < segments) {
        const a = i * (segments + 1) + j, b = a + segments + 1;
        indices.push(a, a + 1, b, b, a + 1, b + 1);
      }
    }
  }
  geometry.setAttribute('position', new Float32BufferAttribute(positions, 3));
  geometry.setAttribute('color', new Float32BufferAttribute(colors, 3));
  geometry.setIndex(indices);
  geometry.computeVertexNormals();
  return geometry;
}

export const point = (p: Setting, lift = 0.07) => new Vector3(p[0], height(p) + lift, p[1]);
