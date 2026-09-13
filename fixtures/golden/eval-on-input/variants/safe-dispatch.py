# Golden fixture: a real function replaces user input — the correct pattern.
# Clean control: must fire zero findings.

import math

ALLOWED = {"sqrt": math.sqrt, "ceil": math.ceil}


def compute(name: str, value: float) -> float:
    return ALLOWED[name](value)
