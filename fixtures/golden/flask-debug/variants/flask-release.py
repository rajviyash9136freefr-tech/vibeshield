import os

from flask import Flask

app = Flask(__name__)
app.secret_key = os.environ["VS_DEMO_SECRET_KEY"]


@app.route("/")
def index():
    return "ok"


if __name__ == "__main__":
    # debug is opt-in via env var — never hardcoded True
    app.run(debug=os.environ.get("FLASK_DEBUG") == "1")
