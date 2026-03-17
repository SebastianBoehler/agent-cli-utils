#!/usr/bin/env python3

import argparse
import asyncio
import json
import plistlib
import sys
import time

import pyatv
from pyatv.const import Protocol
from pyatv.protocols.airplay.player import AirPlayPlayer
from pyatv.protocols.raop.protocols import airplayv2


def build_parser():
    parser = argparse.ArgumentParser()
    parser.add_argument("command", choices=["play"])
    parser.add_argument("--host", required=True)
    parser.add_argument("--credentials", required=True)
    parser.add_argument("--url", required=True)
    parser.add_argument("--hold-seconds", type=float, default=20.0)
    return parser


def print_result(payload):
    print(json.dumps(payload, sort_keys=True), flush=True)


def make_wait_patch(hold_seconds):
    async def patched_wait(self):
        await asyncio.sleep(hold_seconds)

    return patched_wait


def make_play_patch(result):
    async def patched_play_url(self, timing_server_port, url, position=0.0):
        await self._setup_base(timing_server_port)
        await self.start_feedback()
        await self.rtsp.record()

        body = {
            "Content-Location": url,
            "Start-Position-Seconds": position,
            "uuid": self.uuid,
            "streamType": 1,
            "mediaType": "video",
            "rate": 1.0,
        }

        resp = await self.rtsp.connection.post(
            "/play",
            headers=airplayv2.HEADERS,
            body=plistlib.dumps(body, fmt=plistlib.FMT_BINARY),
            allow_error=True,
        )
        result["play_response_code"] = resp.code
        return resp

    return patched_play_url


async def run_play(args):
    loop = asyncio.get_running_loop()
    result = {
        "detail": "",
        "held_seconds": 0.0,
        "mode": "airplay_v2_minimal",
        "ok": False,
        "play_response_code": None,
    }

    original_wait = AirPlayPlayer._wait_for_media_to_end
    original_play = airplayv2.AirPlayV2.play_url
    AirPlayPlayer._wait_for_media_to_end = make_wait_patch(args.hold_seconds)
    airplayv2.AirPlayV2.play_url = make_play_patch(result)

    started = time.monotonic()
    atv = None
    try:
        configs = await pyatv.scan(loop, hosts=[args.host])
        if not configs:
            raise RuntimeError(f"no Apple TV discovered at {args.host}")

        config = configs[0]
        service = config.get_service(Protocol.AirPlay)
        if service is None:
            raise RuntimeError("AirPlay service not found")
        service.credentials = args.credentials

        atv = await pyatv.connect(config, loop)
        await atv.stream.play_url(args.url)

        result["held_seconds"] = round(time.monotonic() - started, 2)
        code = result["play_response_code"]
        if code is None:
            raise RuntimeError("AirPlay play response was not captured")
        result["ok"] = code < 400
        result["detail"] = (
            f"AirPlay /play returned {code}; session held open for "
            f"{result['held_seconds']}s. Playback is not independently verified."
        )
        print_result(result)
        return 0
    except Exception as exc:
        result["held_seconds"] = round(time.monotonic() - started, 2)
        code = result["play_response_code"]
        suffix = f" after /play returned {code}" if code is not None else ""
        result["detail"] = f"{exc}{suffix}"
        print_result(result)
        return 1
    finally:
        AirPlayPlayer._wait_for_media_to_end = original_wait
        airplayv2.AirPlayV2.play_url = original_play
        if atv is not None:
            atv.close()


async def main_async(args):
    if args.command == "play":
        return await run_play(args)
    raise RuntimeError(f"unsupported command {args.command}")


def main():
    args = build_parser().parse_args()
    raise SystemExit(asyncio.run(main_async(args)))


if __name__ == "__main__":
    main()
