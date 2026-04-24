#!/usr/bin/env python3

import argparse
import os

from pyftpdlib.authorizers import DummyAuthorizer
from pyftpdlib.handlers import FTPHandler
from pyftpdlib.servers import FTPServer


def parse_passive_ports(value: str) -> range:
    start_text, end_text = value.split("-", 1)
    start = int(start_text)
    end = int(end_text)
    if start > end:
        raise ValueError("passive port range start must be <= end")
    return range(start, end + 1)


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Start a deterministic local FTP test server"
    )
    parser.add_argument("--host", required=True)
    parser.add_argument("--port", required=True, type=int)
    parser.add_argument("--workspace", required=True)
    parser.add_argument("--user", required=True)
    parser.add_argument("--password", required=True)
    parser.add_argument("--passive-ports", default="30000-30009")
    args = parser.parse_args()

    os.makedirs(args.workspace, exist_ok=True)

    authorizer = DummyAuthorizer()
    authorizer.add_user(args.user, args.password, args.workspace, perm="elradfmwMT")

    handler = FTPHandler
    handler.authorizer = authorizer
    handler.banner = "gofs FTP integration test server ready."
    handler.permit_foreign_addresses = False
    handler.passive_ports = parse_passive_ports(args.passive_ports)
    handler.timeout = 30

    server = FTPServer((args.host, args.port), handler)
    server.max_cons = 32
    server.max_cons_per_ip = 8
    server.serve_forever()


if __name__ == "__main__":
    main()
