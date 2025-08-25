import os
import threading
import webbrowser
from app import app

def open_browser(host: str, port: int):
    url = f"http://{host}:{port}/"
    try:
        webbrowser.open_new(url)
    except Exception:
        pass

host = os.environ.get("HOST", "127.0.0.1")
port = int(os.environ.get("PORT", "5000"))

def env_to_bool(v: str) -> bool:
    return str(v).strip().lower() in {"1", "true", "yes", "on"}

debug = env_to_bool(os.environ.get("DEBUG", "false"))

if (not debug) or (os.environ.get("WERKZEUG_RUN_MAIN") == "true"):
    threading.Thread(target=open_browser, args=(host, port), daemon=True).start()

app.run(host=host, port=port, debug=debug)
