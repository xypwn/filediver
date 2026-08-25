from dds_float16 import DDS
from openexr.types import OpenEXR

import numpy as np
from scipy.spatial.transform import Rotation
from pathlib import Path
from argparse import ArgumentParser

magnitude = 0.1
perturb = [
    Rotation.from_rotvec([magnitude, 0, 0], degrees=True),
    Rotation.from_rotvec([-magnitude, 0, 0], degrees=True),
    Rotation.from_rotvec([0, 0, magnitude], degrees=True),
    Rotation.from_rotvec([0, 0, -magnitude], degrees=True),
    Rotation.from_rotvec([magnitude, 0, magnitude], degrees=True),
    Rotation.from_rotvec([magnitude, 0, -magnitude], degrees=True),
    Rotation.from_rotvec([-magnitude, 0, magnitude], degrees=True),
    Rotation.from_rotvec([-magnitude, 0, -magnitude], degrees=True),
    Rotation.from_rotvec([0, magnitude, 0], degrees=True),
    Rotation.from_rotvec([0, -magnitude, 0], degrees=True),
    Rotation.from_rotvec([magnitude, magnitude, 0], degrees=True),
    Rotation.from_rotvec([magnitude, -magnitude, 0], degrees=True),
    Rotation.from_rotvec([-magnitude, magnitude, 0], degrees=True),
    Rotation.from_rotvec([-magnitude, -magnitude, 0], degrees=True),
]

@np.vectorize(excluded={1, 'pixels'}, signature='(n)->(m)')
def sample_cubemap(vec: np.ndarray, pixels: np.ndarray) -> np.ndarray:
    assert vec.shape[0] == 3, f"Expected 3d vectors got {vec.shape}"
    assert pixels.ndim == 4 and pixels.shape[0] == 6, "Expected cubemap pixels"

    a: float = abs(vec).max()
    vecA: np.ndarray = vec / a

    cubeFaceWidth, cubeFaceHeight = pixels.shape[2], pixels.shape[1]
    if vecA[0] == 1.0 or vecA[0] == -1.0:
        # right / left
        face = 0 if vecA[0] == 1.0 else 1
        offset = -1.0 if vecA[0] == 1.0 else 0.0
        xPixel = int((((vecA[2] + 1.0) / 2.0) + offset) * cubeFaceWidth)
        yPixel = int((((vecA[1] + 1.0) / 2.0)) * cubeFaceHeight)
    elif vecA[1] == 1.0 or vecA[1] == -1.0:
        # up / down
        face = 2 if vecA[1] == 1.0 else 3
        offset = -1.0 if vecA[1] == 1.0 else 0.0
        xPixel = int((((vecA[0] + 1.0) / 2.0)) * cubeFaceWidth)
        yPixel = int((((vecA[2] + 1.0) / 2.0) + offset) * cubeFaceHeight)
    elif vecA[2] == 1.0 or vecA[2] == -1.0:
        # front / back
        face = 4 if vecA[2] == 1.0 else 5
        offset = -1.0 if vecA[2] == -1.0 else 0.0
        xPixel = int((((vecA[0] + 1.0) / 2.0) + offset) * cubeFaceWidth)
        yPixel = int((((vecA[1] + 1.0) / 2.0)) * cubeFaceHeight)

    xPixel = max(min(abs(xPixel), cubeFaceWidth-1), 0)
    yPixel = max(min(abs(yPixel), cubeFaceHeight-1), 0)
    return pixels[face, yPixel, xPixel]

def normalize(vec: np.ndarray) -> np.ndarray:
    return vec / np.sqrt((vec**2).sum())

def make_tile(color: np.ndarray) -> np.ndarray:
    assert color.ndim == 1 and color.shape[0] == 4
    dim = color * 0.25
    return np.array([[color, dim], [dim, color]], dtype=color.dtype)

def build_equirectangular(height: int, width: int, pixels: np.ndarray) -> np.ndarray:
    output = np.zeros((height, width, 4), dtype=pixels.dtype)
    for j in range(output.shape[0]):
        v = 1.0 - float(j) / output.shape[0]
        theta = v * np.pi
        print(f"{j+1}/{output.shape[0]}")

        for i in range(output.shape[1]):
            u = float(i) / output.shape[1]
            phi = u * 2 * np.pi

            x = np.sin(phi) * np.sin(theta) * -1.0
            y = np.cos(theta)
            z = np.cos(phi) * np.sin(theta) * -1.0

            perturbed = np.array([np.array([x, y, z])])
            samples = sample_cubemap(perturbed, pixels)
            # for p in perturb:
            #     samples.append(sample_cubemap(, pixels))
            output[j, i] = np.average(samples, axis=0)
    return output

def main():
    parser = ArgumentParser()
    parser.add_argument("path", type=Path)
    parser.add_argument("output_width", type=int)
    args = parser.parse_args()

    path: Path = args.path

    with path.open("rb") as f:
        pixels = DDS.parse(f).pixels()

    red = np.array([1, 0, 0, 1], dtype=np.float16)
    green = np.array([0, 1, 0, 1], dtype=np.float16)
    blue = np.array([0, 0, 1, 1], dtype=np.float16)
    magenta = np.array([1, 0, 1, 1], dtype=np.float16)
    yellow = np.array([1, 1, 0, 1], dtype=np.float16)
    cyan = np.array([0, 1, 1, 1], dtype=np.float16)

    _ = np.array([
        make_tile(red), # +x
        make_tile(green), # -x
        make_tile(blue), # -y
        make_tile(magenta), # +y
        make_tile(yellow), # +z
        make_tile(cyan), # -z
    ], dtype=np.float16)

    faces = pixels[[0, 1, 5, 4, 3, 2]]
    faces[2] = faces[2,::-1,::-1]
    faces[0] = np.rot90(faces[0], 3)
    faces[1] = np.rot90(faces[1])
    faces[5] = np.rot90(faces[5], 2)
    output_height = args.output_width >> 1

    output = build_equirectangular(output_height, args.output_width, faces)

    exr = OpenEXR.from_pixels(output).serialize()
    exr_path = path.with_suffix(".equirectangular.exr")
    with exr_path.open("wb") as f:
        f.write(exr)

if __name__ == "__main__":
    main()