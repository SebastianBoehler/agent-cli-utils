# Living Room Canvas

A local, on-demand display for voice explanations on the living-room TV.
Native macOS window sharing to Apple TV was confirmed by the viewer on
2 October 2026. Ordinary lesson changes preserve the shared browser window.

## Start and publish

From a fresh checkout, install the command with Python 3 available on PATH:

```sh
cd examples/living-room-canvas
canvas_source="$PWD/cli.py"
mkdir -p "$HOME/.local/bin"
printf '#!/bin/sh\nexec python3 "%s" "$@"\n' "$canvas_source" > "$HOME/.local/bin/living-room-canvas"
chmod +x "$HOME/.local/bin/living-room-canvas"
```

Ensure `~/.local/bin` is on PATH, or invoke `python3 cli.py` directly.
The first start creates the local scene manifest with `gradient.html`.
Publishing changes this gitignored runtime file, keeping the checkout clean.

On this Mac, run:

```sh
living-room-canvas
living-room-canvas status
living-room-canvas show landscape.html
```

The start command runs a local server at `http://127.0.0.1:8765/`.
It does not start at login. Open that address in Chrome and share its window
through macOS Screen Mirroring. See CONNECTION.md for the verified steps.
Leave the outer page open. Publish a lesson again to restart its automatic
sequence without reopening Chrome or reconnecting Apple TV.

```sh
living-room-canvas show gradient.html
living-room-canvas stop
```

Source: `examples/living-room-canvas/` in agent-cli-utils.
Installed command: `~/.local/bin/living-room-canvas`.
Runtime: `/opt/homebrew/bin/python3`; the display helper uses only the Python standard library.
Logs and process ID: `~/.local/state/living-room-canvas/`.
Moving this checkout requires updating the installed command's source path.

## Create the next visual

1. Identify the learner's exact question.
2. Choose the best visual: diagram, plot, animation, 3D surface, or slides.
3. Add the lesson and local assets under `web/`.
4. Inspect the actual rendering and check the maths.
5. Run `living-room-canvas show lesson.html`.
6. Explain aloud and use the learner's response to refine the visual.

The display checks `scene.json` every 500 milliseconds. Publishing replaces
that manifest atomically with a new revision. The iframe loads the new lesson;
the outer shared window stays open. No build, reconnect, or video rendering is
required. Each lesson controls its own automatic pacing.

## Viewer preferences

- One point per slide. Let the visual occupy most of the screen.
- No eyebrow labels, small badges, clocks, or decorative status text.
- Large headings and labels; roughly 36 px or larger at 1080p.
- One short supporting sentence at most. Explain details in the conversation.
- Autoplay immediately, with no required clicks or buttons.
- Hold each stage for 8–12 seconds; move slowly and stop on the final stage.
- Use 3D when two settings and their outcome need to be understood spatially.
- A working display does not prove understanding. Change the explanation when
  the learner remains confused.

## Included toy examples

`landscape.html` shows prediction `w`, target 2, and squared error `(w-2)^2`.
The 3D example uses prediction `a+b` and error `(a+b-2)^2`. Its valley contains
all settings with `a+b=2`. These are teaching examples, not measured neural
network landscapes. The white dot moves into the valley.

`gradient.html` shows descent on `(x-1)^2`, using a step size of 0.2.
The update is `x ← x - 0.2 × 2(x-1)`.

The helper manages the local display only. It does not provide remote machine
access, automatic pairing, or unattended screen-sharing reconnection.

## Rich 3D explanations

The Remotion/Three.js loss-landscape lesson plays directly in the shared browser.
Run `living-room-canvas show motion.html` to replay the installed bundle.
For edits, run `npm run publish` in `tv-motion/`; no MP4 export is needed.
On a fresh checkout, first run `npm ci` and `npm run build` in `tv-motion/`.
The composition requires Node.js and its recorded npm dependencies. See
RENDERING.md for dataset references, setup, teaching semantics and verification.
