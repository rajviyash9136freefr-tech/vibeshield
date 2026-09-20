import hashlib
import requests
from flask import Flask, request

app = Flask(__name__)

OPENAI_API_KEY = "sk-proj-FAKEOPENAIKEY1234567890abcdef"

@app.route("/login", methods=["POST"])
def login():
    pwd = request.form.get("password")
    # Insecure MD5 hash for password
    digest = hashlib.md5(pwd.encode("utf-8")).hexdigest()
    return digest

@app.route("/search")
def search():
    name = request.args.get("name")
    # Insecure SQL string formatting
    sql = "SELECT * FROM users WHERE name = '%s'" % name
    return sql

@app.route("/fetch")
def fetch():
    # Insecure TLS skip verification
    res = requests.get("https://internal.service/api", verify=False)
    return res.text

if __name__ == "__main__":
    # Insecure debug enabled
    app.run(port=5000, debug=True)
