# Verified living-room connection

Recorded on 2 October 2026. Recheck device names and network addresses before
using them again. Device pairing credentials must stay outside this repository.

## Successful path

1. Start `living-room-canvas` on the Mac connected to the home network.
2. Open `http://127.0.0.1:8765/` in Google Chrome.
3. Open macOS Control Center and select Screen Mirroring.
4. Select Apple TV.
5. Select **Fenster oder App** in the source-selection dialog.
6. Select **Fenster oder App wählen**.
7. Choose the Chrome display window using **Dieses Fenster synchronisieren**.
8. Make that browser window full screen.

The viewer confirmed seeing “Hello, living room”, then updated animated lessons
on the Samsung TV through Apple TV. Only the selected browser window is shared.
Do not share the whole desktop for this workflow. The connection interrupted
YouTube; the viewer explicitly permitted that interruption during this session.
Future sessions need their own current authorisation for device actions.

## Native macOS control

On the tested macOS version, menu-bar controls involved MenuBarAgent as well
as ControlCenter. The Computer Use tool could operate Chrome, but did not obtain
usable Control Center content. The user explicitly authorised native Mac APIs.
Native CGEvent keyboard and mouse events then operated the visible controls.
Fn+C opened Control Center. Ctrl+Cmd+F made Chrome full screen.

Inspect a current screenshot before choosing UI targets. Do not reuse absolute
coordinates from this session. Event timing mattered: move the pointer, then
wait approximately 120 milliseconds between mouse down and mouse up.
UI changes can take several seconds to appear in a screenshot.

## Reconnecting after a disconnect

Browser playback alone does not confirm TV playback. Check that Screen Mirroring
shows Apple TV connected and that its current-sharing preview contains the lesson.
After reconnecting, Apple TV can show the macOS source-selection notice until a
window is selected. Use **Inhalte wählen**, then **Fenster oder App** and
**Fenster oder App wählen**. Choose **Dieses Fenster synchronisieren** on the
canvas window. Restart the lesson with `living-room-canvas show motion.html`
after selecting the source, and ask the viewer to confirm the TV image.

Whole-display screenshots worked through `/usr/sbin/screencapture -x -D 1`.
Some development tools and the screenshot skill helper were blocked by an
unaccepted Xcode licence. No licence was accepted by the agent.

## Failed path and device evidence

An AirPlay URL playback test returned HTTP 200, but Apple TV never fetched the
Mac's HLS playlist. The TV showed a spinner and returned to YouTube. A pyatv
playback-status request returned HTTP 500. Do not treat HTTP 200 as visible
playback, and do not repeat that test as the default HTML-display method.
Native window sharing succeeded; re-pairing was not needed in this session.

Discovery at the time found Apple TV 4K generation 3 on `192.168.68.52`,
a living-room HomePod Mini named Wohnzimmer on `192.168.68.53`, and the Mac on
`192.168.68.54`. These addresses can change. Saved pyatv pairing configuration
was in `~/.pyatv.conf`; do not copy or print its secrets.

DIAL can discover and launch supported apps. YouTube search tools find videos.
Neither provides this general-purpose live HTML display. The reusable path
verified here is a local browser window shared through native macOS controls.
Remote agents need authorised access to this Mac before running local tools;
that access was not installed as part of the display setup.
