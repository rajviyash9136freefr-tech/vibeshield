"""Golden fixture: template rendering from request input (category+family anchor,
rule ID reserved for the python template-injection rule). The "rule_id": null
marker below is load-bearing for manifest validation — do not change."""

from flask import Flask, render_template_string, request

app = Flask(__name__)


@app.route("/greet")
def greet():
    name = request.args.get("name", "")
    return render_template_string(f"<p>Hello {name}</p>")  # rule_id: null
