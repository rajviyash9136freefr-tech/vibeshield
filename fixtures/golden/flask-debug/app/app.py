from flask import Flask

app = Flask(__name__)


@app.route("/api/health")
def health():
    return {"status": "ok"}


if __name__ == "__main__":
    app.run(port=5057, debug=True)
