# Golden fixture: GPL-licensed snippet with the copyright header stripped
# (VS-LIC license-missing family). The "rule_id": null marker below is
# load-bearing for manifest validation — do not change.

import socket


def free_port(host="127.0.0.1"):
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.bind((host, 0))
    port = s.getsockname()[1]
    s.close()
    return port  # rule_id: null
