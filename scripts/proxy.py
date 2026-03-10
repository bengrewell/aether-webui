#!/usr/bin/env python3
import argparse
import socket
import threading


BUFFER_SIZE = 65536


def pipe(src: socket.socket, dst: socket.socket) -> None:
    try:
        while True:
            data = src.recv(BUFFER_SIZE)
            if not data:
                break
            dst.sendall(data)
    except Exception:
        pass
    finally:
        try:
            src.shutdown(socket.SHUT_RDWR)
        except Exception:
            pass
        try:
            dst.shutdown(socket.SHUT_RDWR)
        except Exception:
            pass
        try:
            src.close()
        except Exception:
            pass
        try:
            dst.close()
        except Exception:
            pass


def handle_client(client_sock: socket.socket, target_host: str, target_port: int) -> None:
    upstream = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    upstream.connect((target_host, target_port))

    threading.Thread(target=pipe, args=(client_sock, upstream), daemon=True).start()
    threading.Thread(target=pipe, args=(upstream, client_sock), daemon=True).start()


def main() -> None:
    parser = argparse.ArgumentParser(description="Simple threaded TCP proxy")
    parser.add_argument("--listen-host", required=True)
    parser.add_argument("--listen-port", required=True, type=int)
    parser.add_argument("--target-host", required=True)
    parser.add_argument("--target-port", required=True, type=int)
    args = parser.parse_args()

    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server.bind((args.listen_host, args.listen_port))
    server.listen(128)

    print(
        f"Listening on {args.listen_host}:{args.listen_port} "
        f"-> {args.target_host}:{args.target_port}"
    )

    while True:
        client_sock, client_addr = server.accept()
        print(f"Accepted {client_addr[0]}:{client_addr[1]}")
        threading.Thread(
            target=handle_client,
            args=(client_sock, args.target_host, args.target_port),
            daemon=True,
        ).start()


if __name__ == "__main__":
    main()
