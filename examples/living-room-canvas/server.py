"""Local visual display, started on demand."""
from http.server import SimpleHTTPRequestHandler, ThreadingHTTPServer
import json
import os
from pathlib import Path
from urllib.parse import urlsplit

WEB_ROOT = Path(__file__).resolve().parent / "web"
ADDRESS = ("127.0.0.1", 8765)


class Handler(SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=str(WEB_ROOT), **kwargs)

    def do_GET(self):
        if urlsplit(self.path).path != "/health":
            return super().do_GET()
        payload = json.dumps({"service": "living-room-canvas", "pid": os.getpid()}).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(payload)))
        self.end_headers()
        self.wfile.write(payload)

    def end_headers(self):
        self.send_header("Cache-Control", "no-store")
        super().end_headers()

    def log_message(self, *args):
        pass


if __name__ == "__main__":
    with ThreadingHTTPServer(ADDRESS, Handler) as server:
        print("Living Room Canvas listening on http://127.0.0.1:8765/", flush=True)
        server.serve_forever()
