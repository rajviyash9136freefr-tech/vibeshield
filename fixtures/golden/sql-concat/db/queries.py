"""Golden fixture: SQL built by string interpolation (VS-SEC insecure-api family).
The "rule_id": null marker below is load-bearing for manifest validation — do not change."""

import pymysql

conn = pymysql.connect(host="localhost", user="demo", db="demo")


def find_user(cursor, name):
    sql = "SELECT * FROM users WHERE name = '%s'" % name  # rule_id: null
    cursor.execute(sql)
    return cursor.fetchone()
