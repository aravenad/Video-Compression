import sys
import logging
from ui.interface import start_app

# Configure logging
logging.basicConfig(
    filename="app.log",
    filemode="w",
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s"
)

def main():
    try:
        logging.info("Starting Video Compression application")
        start_app()
    except Exception as e:
        logging.error(f"Fatal error: {str(e)}")
        sys.exit(1)
    finally:
        logging.info("Application shutdown")

if __name__ == "__main__":
    main()
