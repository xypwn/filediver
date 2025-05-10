from mmhash2 import murmurhash64a as h64a
from argparse import ArgumentParser

import sys
import os
import select
import tty
import termios
from pathlib import Path

if os.name == 'nt':
    import msvcrt
    import ctypes

    class _CursorInfo(ctypes.Structure):
        _fields_ = [("size", ctypes.c_int),
                    ("visible", ctypes.c_byte)]

def hide_cursor():
    if os.name == 'nt':
        ci = _CursorInfo()
        handle = ctypes.windll.kernel32.GetStdHandle(-11)
        ctypes.windll.kernel32.GetConsoleCursorInfo(handle, ctypes.byref(ci))
        ci.visible = False
        ctypes.windll.kernel32.SetConsoleCursorInfo(handle, ctypes.byref(ci))
    elif os.name == 'posix':
        sys.stdout.write("\033[?25l")
        sys.stdout.flush()

def show_cursor():
    if os.name == 'nt':
        ci = _CursorInfo()
        handle = ctypes.windll.kernel32.GetStdHandle(-11)
        ctypes.windll.kernel32.GetConsoleCursorInfo(handle, ctypes.byref(ci))
        ci.visible = True
        ctypes.windll.kernel32.SetConsoleCursorInfo(handle, ctypes.byref(ci))
    elif os.name == 'posix':
        sys.stdout.write("\033[?25h")
        sys.stdout.flush()

def print_guess(guess: str, index: int):
    if len(guess) != 0 and index < len(guess):
        print(f"\r{guess[:index]}\033[4m{guess[index]}\033[0m{guess[index+1:]}  ", flush=True, end="")
    else:
        print(f"\r{guess}\033[4m \033[0m ", flush=True, end="")

def print_result(guess: str, murmurhash64: int, found: bool, show_not_found: bool, long: bool):
    fmt = f"{murmurhash64 >> 32:08x}" if not long else f"{murmurhash64:016x}"
    if found:
        print(f"\r\033[32m0x{fmt}: found {repr(guess)}\n\033[0m", end="")
    elif show_not_found:
        print(f"\r\033[31m0x{fmt}: not found {repr(guess)}\n\033[0m", end="")

def isData():
    return select.select([sys.stdin], [], [], 0) == ([sys.stdin], [], [])

def handle_control(guess: str, index: int):
    c = sys.stdin.read(2)
    if c == "[D":
        index -= 1
    elif c == "[C":
        index += 1
    else:
        return index
    index = min(max(0, index), len(guess))
    print_guess(guess, index)
    return index

def handle_input(old_guess: str, index: int):
    c = sys.stdin.read(1)
    rehash = False
    if c == '\x1b':         # x1b is ESC
        index = handle_control(old_guess, index)
        return old_guess, index, rehash
    if c == '\n':
        guess = ""
    elif c == '\x7f':
        guess = old_guess
        if index > 0:
            guess = old_guess[:index-1] + old_guess[index:]
        index -= 1
        if index < 0:
            index = 0
        print_guess(guess, index)
    else:
        guess = old_guess[:index] + c + old_guess[index:]
        rehash = True
        index += 1
    return guess, index, rehash

def main():
    parser = ArgumentParser()
    parser.add_argument("hashes_path")
    parser.add_argument("--only-found", "-f", action="store_true", default=False)
    parser.add_argument("--suffixes", "-s", type=Path)
    parser.add_argument("--output", "-o", type=Path)
    args = parser.parse_args()

    found_hashes = set()

    long = False
    show_not_found = not args.only_found
    output = None

    hide_cursor()
    old_settings = termios.tcgetattr(sys.stdin)
    try:
        hashes = []
        with open(args.hashes_path, "r") as f:
            for line in f:
                hashes.append(int(line.strip(), base=16))
                if hashes[-1] > 0xffffffff:
                    long = True

        if args.output:
            output_path: Path = args.output
            output = output_path.open("a")

        suffixes = [""]
        if args.suffixes:
            suffix_path: Path = args.suffixes
            with suffix_path.open("r") as f:
                for line in f:
                    suffixes.append(line.strip())
                    if "/" in line:
                        split = line.strip().split("/")
                        if split[-1] == split[-2]:
                            continue
                        suffixes.append(f"/{split[-1]}/{split[-1]}")

        tty.setcbreak(sys.stdin.fileno())
        rehash = False
        guess = ""
        index = 0
        while 1:
            if rehash:
                for suffix in suffixes:
                    murmurhash64 = h64a(guess + suffix, 0)
                    found = (murmurhash64 in hashes) if long else ((murmurhash64 >> 32) in hashes)
                    if found:
                        found_hashes.add(guess + suffix)
                    print_result(guess + suffix, murmurhash64, found, show_not_found, long)
                print_guess(guess, index)
                rehash = False
            if isData():
                guess, index, rehash = handle_input(guess, index)
    except KeyboardInterrupt:
        print()
    finally:
        show_cursor()
        termios.tcsetattr(sys.stdin, termios.TCSADRAIN, old_settings)
        if output:
            if len(found_hashes) > 0:
                for hash in sorted(list(found_hashes)):
                    _ = output.write(f"{hash}\n")
            output.close()
        elif len(found_hashes) > 0:
            print("Hashes found:")
            for hash in sorted(list(found_hashes)):
                print(f"  {hash}")


if __name__ == "__main__":
    main()
