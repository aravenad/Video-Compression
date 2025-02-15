import logging
import sys
from ui.interface import start_app

logging.basicConfig(
    filename="app.log", filemode="w", level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s"
)

def main():
    logging.info("Starting improved Video Compression application")
    start_app()

if __name__ == "__main__":
    main()
