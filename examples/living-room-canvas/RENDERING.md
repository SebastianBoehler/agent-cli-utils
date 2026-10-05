# Visual production and evidence

## What was inspected

- Published dataset: https://huggingface.co/datasets/sebastianboehler/autoresearch-manim
  The current `cases.jsonl` has 150 accepted examples. Its latent paths,
  embedding clusters, camera choreography and gradient-descent contour examples
  provide useful teaching patterns. The inspected rows are render verified but
  marked `render_quality: low`; successful rendering is not a quality guarantee.
- Local Manim source: `autoresearch_manim_finetune/data/manim_dataset.jsonl`.
  Manim 0.20.1 imports successfully in that project's existing virtual environment.
- Local Remotion source: `autoresearch_remotion_finetune/data/remotion_codex_synthetic_cases.jsonl`
  and `artifacts/candidates/xiaomi-curated-v1/curated_tsx/academic_gradient_descent_curated.tsx`.
  The curated example links a moving marker to a changing model setting.
  Its smaller labels and information cards were not copied into the TV scene.
- Only the Manim dataset was found published under the authenticated personal
  Hugging Face account. Remotion examples were inspected locally; no public
  Remotion dataset ID was verified.

No model was trained. The new composition is original code informed by these
patterns, not an unreviewed dataset completion.

## Choose the rendering method

Use direct SVG/HTML for simple diagrams and plots. Use Manim for precise
mathematical transformations or a finished scientific animation. Use the
Remotion Player and ThreeCanvas for this live 3D explanation: a frame-driven
composition plays directly in the existing browser, without exporting an MP4.

The composition is in `tv-motion/src/`. Geometry, model maths, scene objects,
composition layout and browser playback are separate modules. Each authored
file remains below 300 lines. All motion follows `useCurrentFrame()`; there
are no independent CSS animation clocks or React Three Fiber `useFrame()` loops.

## Publish a changed composition

```sh
cd /Users/sebastianboehler/Documents/GitHub/agent-cli-utils/examples/living-room-canvas/tv-motion
npm ci
npm run publish
```

Run `npm ci` only when dependencies need installing. `npm run publish` checks
TypeScript, builds the local browser bundle, and publishes `motion.html`.
The existing outer window stays shared and full screen. The generated bundle
and node_modules are gitignored; package-lock.json records the dependencies.
The Python display command still starts on demand, with no login service.

For editable timeline review:

```sh
npm run studio
```

Open the URL printed by Studio and choose LossLandscape. The initial preview
ran at http://localhost:3000/LossLandscape. The TV sees a separate player with
no controls, immediate autoplay, no looping, and a stable final frame.

## Teaching contract

The current example explains a loss landscape, not latent space. Model settings
are positions; height represents prediction error. A latent-space plot instead
places learned representations at positions. Do not imply that latent-space
height is inherently an error or that a latent interpolation is gradient descent.

The toy model predicts `(u², v - 0.35u, u)`, with target `(1, 0, 1)`.
Its weighted squared error is:

`L = (u² - 1)² + 0.6(v - 0.35u)² + 0.15(u - 1)²`.

The vertical plot scale is proportional to this actual error. The orange point
uses steps of `settings ← settings - 0.04 × gradient`. Between steps, its height
is recalculated from the current settings; it does not cut through the surface.
The shown sequence lowers error from about 2.66 to 0.16. It does not claim to
have reached the minimum. The full stored trajectory is available in the source.

## Quality checks

- Check derivatives against finite differences.
- Check every update and every displayed frame reduces the error.
- Check the final frame remains stable.
- Run TypeScript and the browser bundle build.
- Inspect first, focal and final frames for clipping, occlusion and text size.
- Inspect actual playback in the shared Chrome window; still rendering alone
  does not establish autoplay or browser GPU performance.

For a still inspection on this Mac, Remotion's headless browser required
`--gl=swangle` to create a WebGL context. This is a frame-inspection setting;
the TV player uses Chrome's normal live rendering.

Remotion's scaffolder was blocked by the existing Xcode/Git licence condition.
The project uses the official local blank template without accepting that licence.
The browser build resolves local npm packages with browser/ESM conditions,
avoiding an unrelated ancestor Yarn Plug'n'Play manifest. Loading Three.js's
Node CommonJS entry caused a browser `process is not defined` error; the
browser/ESM resolution fixes that error rather than adding a process shim.

## Current validation — 2 October 2026

TypeScript and the browser build passed. Finite-difference gradient checks and
checks of all 2160 animated frames passed: the error never increases. First,
focal and final frames were inspected; camera framing and label placement were
adjusted after the review. Live Chrome screenshots confirmed automatic progress
from the first scene to the final stage, with the expected decreasing error.
The final player state was observed without a restart. Physical TV visibility
of this revised lesson was not separately confirmed by the viewer.
