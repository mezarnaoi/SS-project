import logging
from io import BytesIO
from pathlib import PurePath
from typing import Any

import pytesseract
from fastapi import FastAPI, File, HTTPException, UploadFile
from PIL import Image, UnidentifiedImageError

logger = logging.getLogger(__name__)

app = FastAPI(
    title="OCR Sandbox Service",
    description="Isolated OCR service for processing medical document images",
    version="1.0.0",
)

MAX_IMAGE_SIZE_BYTES = 5 * 1024 * 1024

ALLOWED_CONTENT_TYPES = {
    "image/jpeg",
    "image/png",
}

ALLOWED_EXTENSIONS = {
    ".jpg",
    ".jpeg",
    ".png",
}

PIL_FORMAT_TO_CONTENT_TYPE = {
    "JPEG": "image/jpeg",
    "PNG": "image/png",
}


@app.get("/health")
def health() -> dict[str, str]:
    return {
        "status": "ok",
        "service": "ocr-sandbox",
        "engine": "tesseract",
    }


def _validate_filename(filename: str | None) -> None:
    if not filename:
        raise HTTPException(status_code=400, detail="Missing uploaded filename.")

    safe_name = PurePath(filename).name
    extension = PurePath(safe_name).suffix.lower()

    if extension not in ALLOWED_EXTENSIONS:
        raise HTTPException(
            status_code=415,
            detail="Unsupported file extension. Only JPG, JPEG and PNG are allowed.",
        )

    if safe_name != filename:
        raise HTTPException(status_code=400, detail="Invalid uploaded filename.")


def _validate_content_type(content_type: str | None) -> str:
    if content_type not in ALLOWED_CONTENT_TYPES:
        raise HTTPException(
            status_code=415,
            detail="Unsupported file type. Only JPEG and PNG are allowed.",
        )

    return content_type


def _load_verified_image(image_bytes: bytes, expected_content_type: str) -> Image.Image:
    try:
        image = Image.open(BytesIO(image_bytes))
        image.verify()
    except UnidentifiedImageError as exc:
        raise HTTPException(status_code=400, detail="Invalid image file.") from exc
    except Exception as exc:
        raise HTTPException(status_code=400, detail="Could not validate image.") from exc

    try:
        image = Image.open(BytesIO(image_bytes))
        detected_content_type = PIL_FORMAT_TO_CONTENT_TYPE.get(image.format)

        if detected_content_type != expected_content_type:
            raise HTTPException(
                status_code=400,
                detail="Image content does not match the declared file type.",
            )

        return image.convert("RGB")
    except HTTPException:
        raise
    except Exception as exc:
        raise HTTPException(status_code=400, detail="Could not process image.") from exc


def _calculate_average_confidence(ocr_data: dict[str, Any]) -> float:
    confidences: list[float] = []

    for raw_confidence in ocr_data.get("conf", []):
        try:
            confidence = float(raw_confidence)
        except (TypeError, ValueError):
            continue

        if confidence >= 0:
            confidences.append(confidence)

    if not confidences:
        return 0.0

    return sum(confidences) / len(confidences)


def _count_recognized_words(ocr_data: dict[str, Any]) -> int:
    words = ocr_data.get("text", [])

    return sum(1 for word in words if isinstance(word, str) and word.strip())


@app.post("/ocr")
async def extract_text(file: UploadFile = File(...)) -> dict[str, Any]:
    _validate_filename(file.filename)
    expected_content_type = _validate_content_type(file.content_type)

    image_bytes = await file.read()

    if len(image_bytes) == 0:
        raise HTTPException(status_code=400, detail="Empty image payload.")

    if len(image_bytes) > MAX_IMAGE_SIZE_BYTES:
        raise HTTPException(status_code=413, detail="Image payload too large.")

    image = _load_verified_image(image_bytes, expected_content_type)

    try:
        text = pytesseract.image_to_string(image, lang="eng+ron")
        ocr_data = pytesseract.image_to_data(
            image,
            lang="eng+ron",
            output_type=pytesseract.Output.DICT,
        )

        average_confidence = _calculate_average_confidence(ocr_data)

        return {
            "text": text.strip(),
            "confidence": round(average_confidence, 2),
            "recognized_word_count": _count_recognized_words(ocr_data),
            "engine": "tesseract",
        }

    except pytesseract.TesseractError as exc:
        logger.exception("Tesseract OCR engine failed.")
        raise HTTPException(
            status_code=500,
            detail="OCR engine failed while processing the image.",
        ) from exc