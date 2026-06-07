import sys
import struct
from pathlib import Path

# pip install pynacl
from nacl import bindings, exceptions, hash, encoding

MAGIC1 = [
    0x3E8DA00F70BBA701,
    0x994B81AA021F93DB,
    0xAAAF66AAFB37DFB5,
    0xE79639D18E181BF2,
]
MAGIC2 = [
    0x1B3B4D0E8D7A478C,
    0x79A1E77CBFB8B63F,
    0x1776EDE6D312EE2C,
    0x8FF8F2773C453FDF,
]

XK = 0xEB63463F0C116E2C
AK = b"".join(struct.pack("<Q", x ^ XK) for x in MAGIC1)
SK = b"".join(struct.pack("<Q", x ^ XK) for x in MAGIC2)
def nonce_deriver(pk: bytes) -> bytes:
    return hash.blake2b(pk + AK, digest_size=24, encoder=encoding.RawEncoder)
def decode_data(data: bytes) -> bytes | None:
    if len(data) < 48:
        return None
    pk, ct = data[:32], data[32:]
    try:
        return bindings.crypto_box_open_easy(ct, nonce_deriver(pk), pk, SK)
    except exceptions.CryptoError:
        return None
def encode_data(plain: bytes) -> bytes:
    rx_pk = bindings.crypto_scalarmult_base(SK)
    eph_pk, eph_sk = bindings.crypto_box_keypair()
    nonce = nonce_deriver(eph_pk)
    ct = bindings.crypto_box_easy(
        plain,
        nonce,
        rx_pk,
        eph_sk,
    )
    return eph_pk + ct
def decrypt_file(src: Path, dst: Path) -> bool:
    data = src.read_bytes()
    plain = decode_data(data)
    if not plain:
        return False
    dst.write_bytes(plain)
    return True
def main():
    if len(sys.argv) != 2:
        print(f"usage: {sys.argv[0]} <dir>")
        return

    root = Path(sys.argv[1]).resolve()
    if not root.is_dir():
        print("no such dir")
        return

    out_dir = Path(__file__).resolve().parent / "decrypted"
    out_dir.mkdir(exist_ok=True)
    ok = fail = skip = 0
    for p in root.rglob("*"):
        if not p.is_file():
            continue
        if p.suffix not in (".dl_bin", ".dl_typelib"):
            print("[skip] " + str(p))
            skip += 1
            continue
        try:
            out = out_dir / p.name
            if decrypt_file(p, out):
                print("[ok] " + str(p))
                ok += 1
            else:
                print("[fail] " + str(p))
                fail += 1
        except Exception as e:
            print("[FAIL] " + str(p))
            print(f"Error: {e}")
            fail += 1

    print(f"OK: {ok}, FAIL: {fail}, SKIP: {skip}")

if __name__ == "__main__":
    main()
