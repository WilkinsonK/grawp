#!/usr/bin/python3
# Create a tarball of the game world(s).
import argparse, configparser, datetime, itertools, logging, pathlib, zipfile

TARGET_WORLD_NAMES = (
    "",
    "nether",
    "the_end",
)

logging.basicConfig(level=logging.ERROR)
logger = logging.getLogger(__file__)


def archive_world(world_path: pathlib.Path, output_path: pathlib.Path) -> None:
    if not world_path.exists():
        logger.error(f"No such world to create backup {world_path!s}")
        return

    logger.info(f"World {world_path.name!r} exists; Creating backup...")
    logger.debug(f"Backup name: {output_path!s}")
    with zipfile.ZipFile(output_path, mode="w") as zf:
        for root, _, files in world_path.walk():
            for file in map(lambda f: root / f, files):
                logger.debug(f"Compressing file {file!s}...")
                zf.write(file)


def main() -> int:
    ap = argparse.ArgumentParser(
        __file__,
        description="Create a tarbal of the game world(s)."
    )
    ap.add_argument("src", type=pathlib.Path)
    ap.add_argument("dst", type=pathlib.Path, nargs="?", default=".")

    args = ap.parse_args()

    logger.debug("Acquiring arguments from stdin")
    src: pathlib.Path = args.src
    dst: pathlib.Path = args.dst
    logger.debug(f"Source destination: {src!s}")
    logger.debug(f"Output destination: {dst!s}")

    server_properties_path = src / "server.properties"
    if not server_properties_path.exists():
        logger.fatal(f"No such file: {server_properties_path!s}")
        logger.fatal("Exiting")
        return 1

    logger.info("Reading server properties")
    cp = configparser.ConfigParser()
    try:
        with server_properties_path.open() as fd:
            fd = itertools.chain(("[data]",), fd)
            cp.read_file(fd)
    except Exception as e:
        logger.fatal(f"Could not read properties: {e!s}", exc_info=e)
        logger.fatal("Exiting")
        return 2

    logger.info("World identifier found")
    level_name = cp["data"]["level-name"]
    logger.debug(f"World identifier: {level_name!r}")

    current_date = datetime.datetime.now()

    for name in TARGET_WORLD_NAMES:
        name = "_".join((level_name, name)).removesuffix("_")
        server_world_path = src / name
        output_name = f"{src}-{name}.world.{current_date.month:02}{current_date.day:02}{current_date.year:04}.zip"
        output_path = dst / output_name
        archive_world(server_world_path, output_path)

    return 0


if __name__ == "__main__":
    exit(main())
