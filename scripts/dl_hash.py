def dlhash(text: str) -> int:
    val = 5381
    for char in text:
        val = (((val * 33) & 0xffffffff) + ord(char)) & 0xffffffff
    return (val - 5381) & 0xffffffff

def swap_endianness(val: int) -> int:
    return ((val & 0xff) << 24) | ((val >> 8 & 0xff) << 16) | ((val >> 16 & 0xff) << 8) | (val >> 24 & 0xff)

from argparse import ArgumentParser

def main():
    parser = ArgumentParser()
    parser.add_argument("text", nargs="+")
    parser.add_argument("--swap", action="store_true")
    args = parser.parse_args()

    for text in args.text:
        val = dlhash(text)
        if args.swap:
            val = swap_endianness(val)
            print(f"{text} = 0x{val:08x},")
        else:
            print(f"{text} = 0x{val:08x},")

if __name__ == "__main__":
    main()