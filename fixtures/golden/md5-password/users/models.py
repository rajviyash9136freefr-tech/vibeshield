"""Golden fixture: MD5 used for password storage (VS-SEC insecure-api family).
The "rule_id": null marker below is load-bearing for manifest validation — do not change.
SQL here is parameterized on purpose: this file must fire exactly one finding."""

import hashlib

import pymysql

conn = pymysql.connect(host="localhost", user="demo", db="demo")


def store_password(cursor, username, password):
    digest = hashlib.md5(password.encode("utf-8")).hexdigest()  # rule_id: null
    sql = "INSERT INTO users (name, pass) VALUES (%s, %s)"
    cursor.execute(sql, (username, digest))
