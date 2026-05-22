from io import BytesIO
from typing import Any

import pytesseract
from fastapi import FastAPI, File, HTTPException, UploadFile
from PIL import Image, UnidentifiedImageError

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


@app.get("/health")
def health() -> dict[str, str]:
    return {
        "status": "ok",
        "service": "ocr-sandbox",
        "engine": "tesseract",
    }


@app.post("/ocr")
async def extract_text(file: UploadFile = File(...)) -> dict[str, Any]:
    if file.content_type not in ALLOWED_CONTENT_TYPES:
        raise HTTPException(
            status_code=415,
            detail="Unsupported file type. Only JPEG and PNG are allowed.",
        )

    image_bytes = await file.read()

    if len(image_bytes) == 0:
        raise HTTPException(status_code=400, detail="Empty image payload.")

    if len(image_bytes) > MAX_IMAGE_SIZE_BYTES:
        raise HTTPException(status_code=413, detail="Image payload too large.")

    try:
        image = Image.open(BytesIO(image_bytes))
        image.verify()
    except UnidentifiedImageError as exc:
        raise HTTPException(status_code=400, detail="Invalid image file.") from exc
    except Exception as exc:
        raise HTTPException(status_code=400, detail="Could not validate image.") from exc

    image = Image.open(BytesIO(image_bytes)).convert("RGB")

    try:
        text = pytesseract.image_to_string(image, lang="eng+ron")

        ocr_data = pytesseract.image_to_data(
            image,
            lang="eng+ron",
            output_type=pytesseract.Output.DICT,
        )

        confidences: list[float] = []
        for raw_confidence in ocr_data.get("conf", []):
            try:
                confidence = float(raw_confidence)
            except (TypeError, ValueError):
                continue

            if confidence >= 0:
                confidences.append(confidence)

        average_confidence = (
            sum(confidences) / len(confidences)
            if len(confidences) > 0
            else 0.0
        )

        return {
            "text": text.strip(),
            "confidence": round(average_confidence, 2),
            "engine": "tesseract",
        }

    except pytesseract.TesseractError as exc:
        raise HTTPException(status_code=500, detail="OCR engine failed.") from exc