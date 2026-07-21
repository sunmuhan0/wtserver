#!/usr/bin/env python3
"""
StatShark Browser Proxy
- Keeps Chrome alive via DrissionPage (non-headless + Xvfb)
- Proxies API requests to statshark.net through the browser
- Automatically handles Turnstile token from localStorage
"""

import json
import os
import sys
import time
import threading
import traceback
from http.server import HTTPServer, BaseHTTPRequestHandler
from socketserver import ThreadingMixIn

os.environ.setdefault("DISPLAY", ":99")

PORT = int(os.environ.get("PROXY_PORT", "8082"))
SS_BASE = "https://statshark.net"
CHROME_PATH = "/usr/bin/google-chrome-stable"

page = None
page_lock = threading.Lock()
ready = False
token = ""
cf_clearance = ""


def init_browser():
    global page, ready, token, cf_clearance
    from DrissionPage import ChromiumPage, ChromiumOptions

    co = ChromiumOptions()
    co.set_browser_path(CHROME_PATH)
    co.set_argument("--no-sandbox")
    co.set_argument("--disable-dev-shm-usage")
    co.set_argument("--disable-gpu")
    co.set_argument("--disable-blink-features=AutomationControlled")

    print("[proxy] starting Chrome...", flush=True)
    with page_lock:
        page = ChromiumPage(co)

    print("[proxy] navigating to statshark.net...", flush=True)
    page.get(f"{SS_BASE}/")

    print("[proxy] waiting for Turnstile to solve (25s)...", flush=True)
    time.sleep(25)

    _refresh_credentials()
    ready = True
    print(f"[proxy] ready: token={'yes' if token else 'no'}, cf_clearance={'yes' if cf_clearance else 'no'}", flush=True)


def _refresh_credentials():
    global token, cf_clearance
    with page_lock:
        t = page.run_js("return localStorage.getItem('turnstile_token') || ''")
        if t:
            token = t

        cookies = page.cookies()
        for c in cookies:
            if c.get("name") == "cf_clearance":
                cf_clearance = c.get("value", "")

    print(f"[proxy] credentials: token_len={len(token)}, cf_clearance={'yes' if cf_clearance else 'no'}", flush=True)


def _refresh_session():
    global ready, token, cf_clearance
    print("[proxy] refreshing session...", flush=True)
    ready = False
    token = ""
    cf_clearance = ""

    with page_lock:
        page.get(f"{SS_BASE}/")

    print("[proxy] waiting for Turnstile (25s)...", flush=True)
    time.sleep(25)

    _refresh_credentials()
    ready = True
    print("[proxy] session refreshed", flush=True)


def _browser_fetch(method, url, headers, body):
    with page_lock:
        token_val = token
        headers_js = ""
        for k, v in headers.items():
            headers_js += f"xhr.setRequestHeader({json.dumps(k)}, {json.dumps(v)});\n"

        body_arg = json.dumps(body) if body else "''"

        page.run_js(f"window.__proxyResult = ''")
        page.run_js(f"""var xhr = new XMLHttpRequest();
xhr.open({json.dumps(method)}, {json.dumps(url)}, false);
{headers_js}
try {{ xhr.send({body_arg}); window.__proxyResult = xhr.status + '|||' + xhr.responseText.substring(0, 1048576); }} catch(e) {{ window.__proxyResult = 'ERR|||' + e.message; }}""")
        result = page.run_js("return window.__proxyResult")

    if not result or not isinstance(result, str):
        return {"status": 0, "body": ""}

    if result.startswith("ERR|||"):
        return {"status": 0, "body": result[6:]}

    parts = result.split("|||", 1)
    try:
        status = int(parts[0])
    except ValueError:
        status = 0
    body = parts[1] if len(parts) > 1 else ""
    return {"status": status, "body": body}


class ProxyHandler(BaseHTTPRequestHandler):
    def log_message(self, format, *args):
        pass

    def do_GET(self):
        self._proxy("GET")

    def do_POST(self):
        self._proxy("POST")

    def _proxy(self, method):
        global token

        if not ready:
            self._respond(503, {"error": "browser not ready"})
            return

        if self.path == "/health":
            self._respond(200, {"ready": ready, "token_len": len(token), "cf_clearance": "yes" if cf_clearance else "no"})
            return

        if self.path == "/refresh":
            threading.Thread(target=_refresh_session, daemon=True).start()
            self._respond(200, {"status": "refreshing"})
            return

        content_length = 0
        req_body = ""
        if method == "POST":
            content_length = int(self.headers.get("Content-Length", 0))
            if content_length > 0:
                req_body = self.rfile.read(content_length).decode("utf-8")

        headers = {
            "Accept": "application/json, text/plain, */*",
            "X-Turnstile-Token": token,
            "Referer": f"{SS_BASE}/players",
        }

        if method == "POST":
            headers["Content-Type"] = "application/json"
            headers["Origin"] = SS_BASE

        url = f"{SS_BASE}{self.path}"

        try:
            result = _browser_fetch(method, url, headers, req_body)
        except Exception as e:
            traceback.print_exc()
            self._respond(500, {"error": str(e)})
            return

        status = result.get("status", 0)
        body = result.get("body", "")

        if status == 406:
            print("[proxy] got 406, refreshing session...", flush=True)
            threading.Thread(target=_refresh_session, daemon=True).start()
            self._respond(406, {"error": "token expired, refreshing"})
            return

        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(body.encode("utf-8"))

    def _respond(self, code, data):
        body = json.dumps(data)
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(body.encode("utf-8"))


class ThreadedHTTPServer(ThreadingMixIn, HTTPServer):
    daemon_threads = True


def main():
    init_browser()

    server = ThreadedHTTPServer(("127.0.0.1", PORT), ProxyHandler)
    print(f"[proxy] listening on 127.0.0.1:{PORT}", flush=True)
    server.serve_forever()


if __name__ == "__main__":
    main()
