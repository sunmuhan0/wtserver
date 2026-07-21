#!/usr/bin/env python3
"""Fetch Cloudflare Turnstile token from statshark.net using DrissionPage + Xvfb."""

import json
import os
import subprocess
import sys
import time

from DrissionPage import ChromiumPage, ChromiumOptions

TOKEN_FILE = "/root/project/wtserver/token.json"
SS_PLAYER_URL = "https://statshark.net/player/224501637"
TIMEOUT = 120


def start_xvfb():
    subprocess.Popen(
        ["Xvfb", ":99", "-screen", "0", "1920x1080x24", "-ac"],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )
    os.environ["DISPLAY"] = ":99"
    time.sleep(1)


def main():
    start_xvfb()

    co = ChromiumOptions()
    co.set_browser_path("/usr/bin/google-chrome-stable")
    co.set_argument("--no-sandbox")
    co.set_argument("--disable-dev-shm-usage")
    co.set_argument("--disable-blink-features=AutomationControlled")
    co.set_argument("--window-size=1920,1080")
    co.set_user_agent(
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
        "(KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36"
    )

    print("[fetch] starting Chrome (non-headless via Xvfb)...")
    page = ChromiumPage(co)

    try:
        print(f"[fetch] navigating to {SS_PLAYER_URL}...")
        page.get(SS_PLAYER_URL)
        time.sleep(5)
        print(f"[fetch] page title: {page.title}")

        print("[fetch] waiting for Turnstile to render and solve...")
        start = time.time()
        turnstile_token = ""
        cf_clearance = ""

        while time.time() - start < TIMEOUT:
            elapsed = int(time.time() - start)

            try:
                val = page.run_js(
                    """document.querySelector('input[name="cf-turnstile-response"]')?.value || ''"""
                )
                if val:
                    turnstile_token = val
                    print(f"[fetch] got turnstile token at {elapsed}s (len={len(val)})")
                    break
            except Exception:
                pass

            if elapsed > 0 and elapsed % 15 == 0:
                try:
                    iframes = page.run_js("document.querySelectorAll('iframe').length")
                    dialogs = page.run_js(
                        "document.querySelectorAll('.turnstile-dialog').length"
                    )
                    print(f"[fetch] poll {elapsed}s: iframes={iframes}, dialogs={dialogs}")
                except Exception:
                    print(f"[fetch] poll {elapsed}s")

            time.sleep(3)

        cookies = page.cookies()
        for c in cookies:
            if c.get("name") == "cf_clearance":
                cf_clearance = c.get("value", "")
                print(f"[fetch] got cf_clearance (len={len(cf_clearance)})")

        if turnstile_token or cf_clearance:
            data = {
                "turnstile_token": turnstile_token,
                "cf_clearance": cf_clearance,
                "updated_at": time.strftime("%Y-%m-%dT%H:%M:%S%z"),
            }
            with open(TOKEN_FILE, "w") as f:
                json.dump(data, f, indent=2)
            print(f"[fetch] saved to {TOKEN_FILE}")
            print(f"  turnstile: {'yes' if turnstile_token else 'no'}")
            print(f"  cf_clearance: {'yes' if cf_clearance else 'no'}")
            return 0
        else:
            print("[fetch] failed to get token")
            return 1

    finally:
        page.quit()


if __name__ == "__main__":
    sys.exit(main())
