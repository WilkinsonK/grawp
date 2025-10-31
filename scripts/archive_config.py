#!/usr/bin/python3
# Creates a backup of the server configuration(s).
import argparse, datetime, logging, pathlib, zipfile

TARGET_FILE_NAMES = [
    "banned-ips.json",
    "banned-players.json",
    "ops.json",
    "server.properties",
    "usercache.json",
    "whitelist.json"
]

logging.basicConfig(level=logging.ERROR)
logger = logging.getLogger(__file__)


def main() -> int:
    ap = argparse.ArgumentParser(
        __file__,
        description="Create a backup of the game world(s)."
    )
    ap.add_argument("src", type=pathlib.Path)
    ap.add_argument("dst", type=pathlib.Path, nargs="?", default=".")

    args = ap.parse_args()

    logger.debug("Acquiring arguments from stdin")
    src: pathlib.Path = args.src
    dst: pathlib.Path = args.dst
    logger.debug(f"Source destination: {src!s}")
    logger.debug(f"Output destination: {dst!s}")

    logger.info("Creating backup...")
    current_date = datetime.datetime.now()
    output_name = f"{src}.config.{current_date.month:02}{current_date.day:02}{current_date.year:04}.zip"
    output_path = dst / output_name
    logger.debug(f"Backup name: {output_path!s}")
    with zipfile.ZipFile(output_path, mode="w") as zf:
        for file in map(lambda f: src / f, TARGET_FILE_NAMES):
            if not file.exists():
                continue
            logger.debug(f"Compressing file {file!s}...")
            zf.write(file)

    return 0


if __name__ == "__main__":
    exit(main())
