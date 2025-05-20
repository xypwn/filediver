from argparse import ArgumentParser

import sys
import os
from pathlib import Path

if os.name == 'nt':
    import msvcrt
    import ctypes
    from ctypes.wintypes import DWORD, BOOL, HANDLE
    def h64a(key: str, seed: int) -> int:
        MIX = 0xc6a4a7935bd1e995
        SHIFTS = 47

        b = key.encode()
        hash = (seed ^ (len(b) * MIX)) & 0xFFFFFFFFFFFFFFFF

        while len(b) >= 8:
            value = b[0] | b[1] << 8 | b[2] << 16 | b[3] << 24 | b[4] << 32 | b[5] << 40 | b[6] << 48 | b[7] << 56
            b = b[8:]

            value = (value * MIX) & 0xFFFFFFFFFFFFFFFF
            value ^= value >> SHIFTS
            value = (value * MIX) & 0xFFFFFFFFFFFFFFFF

            hash = hash ^ value
            hash = (hash * MIX) & 0xFFFFFFFFFFFFFFFF

        if len(b) > 0:
            for i, value in enumerate(b):
                hash ^= (value << (8 * i)) & 0xFFFFFFFFFFFFFFFF
            hash = (hash * MIX) & 0xFFFFFFFFFFFFFFFF

        hash ^= hash >> SHIFTS

        hash = (hash * MIX) & 0xFFFFFFFFFFFFFFFF
        hash ^= hash >> SHIFTS

        return hash

    class _CursorInfo(ctypes.Structure):
        _fields_ = [("size", ctypes.c_int),
                    ("visible", ctypes.c_byte)]

    # All input events are redirected to stdin as ansii codes
    ENABLE_VIRTUAL_TERMINAL_INPUT = 0x0200
    # Enable mouse input events
    ENABLE_MOUSE_INPUT = 0x0010
    # Need to be able to enable mouse events
    ENABLE_EXTENDED_FLAGS = 0x0080
else:
    import tty
    import termios
    import select
    from mmhash2 import murmurhash64a as h64a

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

def handle_from_file(file):
    if os.name != 'nt':
        return file
    filenum = None
    match file:
        case sys.stdin:
            filenum = -10
        case sys.stdout:
            filenum = -11
        case sys.stderr:
            filenum = -12
        case _:
            print(f"Unknown file {file}")
            return None
    return HANDLE(ctypes.windll.kernel32.GetStdHandle(filenum))

def get_settings(file):
    if os.name == 'posix':
        return termios.tcgetattr(file)
    elif os.name == 'nt':
        handle = handle_from_file(file)
        mode = DWORD()
        ctypes.windll.kernel32.GetConsoleMode(handle, ctypes.byref(mode))
        return mode
    return None

def set_settings(file, settings):
    if os.name == 'posix':
        termios.tcsetattr(sys.stdin, termios.TCSADRAIN, settings)
    elif os.name == 'nt':
        handle = handle_from_file(file)
        ctypes.windll.kernel32.SetConsoleMode(handle, settings)

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
    if os.name == 'nt':
        handle = handle_from_file(sys.stdin)
        count = DWORD(1)
        wait_all = BOOL(0)
        wait_millis = DWORD(0xFFFFFFFF) #INFINITE == 0xFFFFFFFF
        return ctypes.windll.kernel32.SetConsoleMode(count, ctypes.byref(handle), wait_all, wait_millis) == 0
    elif os.name == 'posix':
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
    if c == '\x03':
        raise KeyboardInterrupt
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
    parser.add_argument("--force-long", "-l", action="store_true", default=False)
    args = parser.parse_args()

    found_hashes = set()

    long = args.force_long
    show_not_found = not args.only_found
    output = None

    hide_cursor()
    old_settings = get_settings(sys.stdin)
    try:
        hashes = []
        with open(args.hashes_path, "r") as f:
            for line in f:
                hashes.append(int(line.strip(), base=16))
                if hashes[-1] > 0xffffffff and not long:
                    long = True
        hashes = set(hashes)
        if args.output:
            output_path: Path = args.output
            output = output_path.open("a")

        suffixes = [""]
        if args.suffixes:
            suffix_path: Path = args.suffixes
            with suffix_path.open("r") as f:
                for line in f:
                    suffixes.append(line.strip())
                    suffixes.append(f"/{line.strip()}")
                    if "/" in line:
                        split = line.strip().split("/")
                        if split[-1] == split[-2]:
                            continue
                        suffixes.append(f"/{split[-1]}/{split[-1]}")
                        if "n_units" in split[-2]:
                            suffixes.append(f"/terrain_units/{split[-1]}")
                            suffixes.append(f"terrain_units/{split[-1]}")
                            suffixes.append(f"_terrain_units/{split[-1]}")
                        split2 = split[-1].split("_")
                        if split2[-1].isnumeric():
                            suffixes.append(f"/{'_'.join(split2[:-1])}/{split[-1]}")
                        if split2[0] == "cy":
                            suffixes.append(f"/{'_'.join(['cyborg'] + split2[1:])}/{split[-1]}")
                        if split2[0] == "il":
                            suffixes.append(f"/{'_'.join(['illuminate'] + split2[1:])}/{split[-1]}")
                    else:
                        suffixes.append(f"/{line.strip()}/{line.strip()}")

        if os.name == 'posix':
            tty.setcbreak(sys.stdin.fileno())
        elif os.name == 'nt':
            set_settings(sys.stdin, ENABLE_EXTENDED_FLAGS | ENABLE_VIRTUAL_TERMINAL_INPUT)
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
        set_settings(sys.stdin, old_settings)
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
