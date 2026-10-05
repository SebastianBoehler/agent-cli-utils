# Living Room Canvas

This is a working local prototype for voice explanations shown on a TV.
Read README.md and CONNECTION.md before reconnecting devices.
Read RENDERING.md before editing the Remotion composition.

## Explanation workflow

- Show one point per slide, with the visual taking most of the screen.
- No eyebrow labels, small badges, clocks, or decorative status text.
- Use large text and plot labels, roughly 36 px or larger at 1080p.
- Use at most one short supporting sentence; explain details aloud.
- Recover the learner's exact confusion before choosing the visual.
- Choose a diagram, plot, animation, 3D surface, or slide sequence for the point
  being explained. Do not reuse one plot merely because it already exists.
- Use contrast, local highlights, moving points and a meaningful trail to direct
  attention. Keep the camera steady unless a move reveals the concept.
- Keep geometry in frame and make the focal point and arrows visible. Inspect
  first, focal and final frames; a successful build is not visual approval.
- Distinguish latent representations from model settings and loss landscapes.
- Use 3D when two changing quantities and an outcome are central to the idea.
  Name each axis and preserve the meaning of positions, height, colour, and time.
- Start with a concrete example and make cause and effect visible. Distinguish
  illustrative toy examples from actual model behaviour or measurements.
- Playback must begin automatically. The viewer is speaking from the couch and
  cannot click buttons. Hold stages for roughly 8–12 seconds, move slowly, and
  stop at the final stage rather than repeatedly restarting the explanation.
- A successful TV connection does not establish that the learner understands.
  If the learner remains confused, change the explanation and visual approach.

## Fast update loop

1. Run `living-room-canvas status`, or `living-room-canvas` to start on demand.
2. Keep the existing shared browser window and display page open.
3. Create or edit a self-contained lesson under web/, with local assets.
4. Check the maths and inspect the rendered lesson.
5. Run `living-room-canvas show lesson.html` to replace the embedded lesson.
   For Remotion edits, run `npm run publish` in tv-motion/ to check, build and show.
6. Explain the change in the voice conversation and check the learner's response.

Publishing changes the inner lesson, preserving the shared browser window and
its full-screen state. Do not reconnect AirPlay for ordinary visual updates.
Do not add a mandatory click or reload the outer display page.

## Scope

The installed helper manages a loopback web server. It does not automatically
pair TVs, connect AirPlay, provide remote computer access, or start at login.
Device actions require current user authorisation; saved notes do not grant it.
