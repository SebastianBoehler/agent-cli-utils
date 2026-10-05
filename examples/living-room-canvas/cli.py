"""Manage the local canvas without changing TV sharing or login settings."""
import argparse
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import time
from urllib.error import URLError
from urllib.request import urlopen
import uuid

ROOT = Path(__file__).resolve().parent
WEB_ROOT = ROOT / "web"
STATE = Path.home() / ".local/state/living-room-canvas"
URL = "http://127.0.0.1:8765/"


def health():
    with urlopen(URL + "health", timeout=1) as response:
        result = json.load(response)
    if result.get("service") != "living-room-canvas":
        raise ValueError("Port 8765 belongs to another service.")
    return result


def start():
    manifest = WEB_ROOT / "scene.json"
    if not manifest.exists():
        manifest.write_text(json.dumps({"path": "gradient.html", "revision": "initial"}))
    try:
        result = health()
    except URLError:
        result = None
    if result:
        print("Already running:", URL)
        return
    STATE.mkdir(parents=True, exist_ok=True, mode=0o700)
    log_path = STATE / "server.log"
    with log_path.open("a") as log:
        process = subprocess.Popen(
            [sys.executable, str(ROOT / "server.py")],
            stdin=subprocess.DEVNULL, stdout=log, stderr=log,
            start_new_session=True,
        )
    for _ in range(30):
        if process.poll() is not None:
            raise RuntimeError(f"Display startup failed. Read {log_path}.")
        try:
            result = health()
        except URLError:
            time.sleep(0.1)
            continue
        if result["pid"] != process.pid:
            process.terminate()
            raise RuntimeError("Another process acquired the display port.")
        (STATE / "server.pid").write_text(str(process.pid))
        print("Started:", URL)
        print("Open this page on the Mac and share its window to Apple TV.")
        return
    process.terminate()
    raise RuntimeError(f"Display did not become ready. Read {log_path}.")


def show(path):
    scene = (WEB_ROOT / path).resolve()
    try:
        relative = scene.relative_to(WEB_ROOT)
    except ValueError:
        raise ValueError("The visual must be inside the canvas web directory.")
    if not scene.is_file() or scene.suffix != ".html" or relative.name == "index.html":
        raise ValueError("Choose an existing visual HTML file, not index.html.")
    health()
    manifest = WEB_ROOT / "scene.json"
    temporary = manifest.with_suffix(".tmp")
    temporary.write_text(json.dumps({
        "path": relative.as_posix(), "revision": str(uuid.uuid4()),
    }))
    temporary.replace(manifest)
    print("Published to the existing shared window:", relative.as_posix())


def stop():
    result = health()
    pid = int((STATE / "server.pid").read_text())
    if result["pid"] != pid:
        raise RuntimeError("The running service does not match the saved process.")
    os.kill(pid, signal.SIGTERM)
    (STATE / "server.pid").unlink()
    print("Stopped the local display server.")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest="command")
    for name in ("start", "status", "stop"):
        commands.add_parser(name)
    publish = commands.add_parser("show")
    publish.add_argument("html", type=Path)
    args = parser.parse_args()
    try:
        if args.command in (None, "start"):
            start()
        elif args.command == "show":
            show(args.html)
        elif args.command == "stop":
            stop()
        else:
            print(json.dumps(health()))
    except (OSError, ValueError, RuntimeError, URLError) as error:
        parser.exit(1, f"living-room-canvas: {error}\n")


if __name__ == "__main__":
    main()
