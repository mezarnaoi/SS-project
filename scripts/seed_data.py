#!/usr/bin/env python3

import base64
import os
import random
import sys
from datetime import datetime, timedelta, timezone

import pymongo
from cryptography.hazmat.primitives.ciphers.aead import AESGCM

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
KEY_FILE = os.path.join(SCRIPT_DIR, "..", "secrets", "db_encryption.key")

MONGO_URI = "mongodb://admin:abcdef@localhost:27019/"
DB_NAME = "mqtt-streaming-server"
COLLECTION_NAME = "photos"


def load_key() -> bytes:
    raw = None
    if os.path.exists(KEY_FILE):
        with open(KEY_FILE, "r") as f:
            raw = f.read().strip()
    else:
        raw = os.environ.get("DB_ENCRYPTION_KEY", "")

    if not raw:
        sys.exit(
            "ERROR: encryption key not found.\n"
            f"  Run ./scripts/gen-db-key.sh or set DB_ENCRYPTION_KEY.\n"
            f"  Expected: {KEY_FILE}"
        )

    try:
        key = base64.b64decode(raw)
    except Exception:
        key = raw.encode()

    if len(key) not in (16, 24, 32):
        sys.exit(f"ERROR: key must be 16, 24 or 32 bytes after base64 decode, got {len(key)}")

    return key


def encrypt(key: bytes, plaintext: str) -> str:
    """AES-256-GCM encrypt. Matches Go utils.EncryptString exactly:
    output = base64(nonce || ciphertext+tag), empty string passes through."""
    if not plaintext:
        return plaintext
    nonce = os.urandom(12)
    aesgcm = AESGCM(key)
    ciphertext = aesgcm.encrypt(nonce, plaintext.encode(), None)
    return base64.b64encode(nonce + ciphertext).decode()


PHI_FIELDS = [
    "text",
    "unitate_medicala",
    "adresa_unitate_medicala",
    "telefon_unitate_medicala",
    "numar_fisa",
    "societate_unitate",
    "adresa_angajator",
    "telefon_angajator",
    "nume",
    "prenume",
    "cnp",
    "profesie_functie",
    "loc_de_munca",
    "recomandari",
]

NAMES = ["Ion", "Maria", "Andrei", "Elena", "Radu", "Ana", "George", "Ioana",
         "Mihai", "Cristina", "Alexandru", "Gabriela", "Florin", "Daniela", "Vlad"]
SURNAMES = ["Popescu", "Ionescu", "Dumitru", "Stoica", "Radu", "Gheorghe", "Matei",
            "Florea", "Costea", "Marinescu", "Dinu", "Toma", "Stanciu"]
JOBS = ["Inginer", "Programator", "Medic", "Profesor", "Contabil",
        "Sofer", "Manager", "Asistent", "Operator", "Electrician", "Mecanic"]
COMPANIES = [
    ("TechCorp SRL", "Bd. Unirii nr 10", "0211111111"),
    ("MediSoft SA", "Str. Victoriei nr 5", "0212222222"),
    ("AutoProd SRL", "Calea Grivitei nr 20", "0213333333"),
    ("EduPlus SRL", "Str. Scolii nr 3", "0214444444"),
    ("BuildMax SA", "Bd. Industriilor nr 15", "0215555555"),
]
MEDICAL_UNITS = [
    ("Clinica Sanatatea", "Str. Sanatatii nr 1", "0700100200"),
    ("Centrul Medical Vida", "Bd. Unirii nr 45", "0700200300"),
]


def generate_photo(key: bytes, days_ago=None, expiring_next_month=False) -> dict:
    now = datetime.now(timezone.utc)

    if days_ago is not None:
        timestamp = now - timedelta(days=days_ago)
    else:
        timestamp = now - timedelta(days=random.randint(0, 365))

    if expiring_next_month:
        data_urm = now + timedelta(days=random.randint(1, 30))
    else:
        data_urm = timestamp + timedelta(days=random.randint(180, 730))

    control_types = ["Angajare", "Periodic", "Adaptare", "Reluare", "Supraveghere", "Alte"]
    selected_control = random.choice(control_types)

    aviz_types = ["APT", "APT CONDITIONAT", "INAPT TEMPORAR", "INAPT"]
    selected_aviz = random.choices(aviz_types, weights=[70, 15, 10, 5], k=1)[0]

    company = random.choice(COMPANIES)
    unit = random.choice(MEDICAL_UNITS)
    nume = random.choice(SURNAMES)
    prenume = random.choice(NAMES)

    ocr_conf = random.uniform(70.0, 99.9)
    needs_review = ocr_conf < 95.0

    doc = {
        # --- non-PHI (plaintext) ---
        "timestamp": timestamp,
        "image_type": "jpeg",
        "device_id": f"device-{random.randint(1, 5)}",
        "tip_control": f"Control {selected_control}",
        "control_angajare": selected_control == "Angajare",
        "control_periodic": selected_control == "Periodic",
        "control_adaptare": selected_control == "Adaptare",
        "control_reluare": selected_control == "Reluare",
        "control_supraveghere": selected_control == "Supraveghere",
        "control_alte": selected_control == "Alte",
        "aviz_medical": selected_aviz,
        "aviz_apt": selected_aviz == "APT",
        "aviz_apt_conditionat": selected_aviz == "APT CONDITIONAT",
        "aviz_inapt_temporar": selected_aviz == "INAPT TEMPORAR",
        "aviz_inapt": selected_aviz == "INAPT",
        "data": timestamp,
        "data_urm_examinari": data_urm,
        "ocr_confidence": ocr_conf,
        "needs_review": needs_review,
        "processing_time_ms": random.randint(200, 1800),
        "reviewed_by": "admin" if needs_review and random.choice([True, False]) else None,

        # --- PHI (will be encrypted below) ---
        "text": f"OCR: {nume} {prenume}",
        "unitate_medicala": unit[0],
        "adresa_unitate_medicala": unit[1],
        "telefon_unitate_medicala": unit[2],
        "numar_fisa": f"FISA-{random.randint(1000, 9999)}",
        "societate_unitate": company[0],
        "adresa_angajator": company[1],
        "telefon_angajator": company[2],
        "nume": nume,
        "prenume": prenume,
        "cnp": (f"{random.randint(1,2)}{random.randint(50,99)}"
                f"{random.randint(10,12)}{random.randint(10,28)}123456"),
        "profesie_functie": random.choice(JOBS),
        "loc_de_munca": company[0],
        "recomandari": "Nicio recomandare" if selected_aviz == "APT" else "Reevaluare necesara",
    }

    for field in PHI_FIELDS:
        doc[field] = encrypt(key, doc[field])

    return doc


def seed_data():
    key = load_key()
    print(f"Loaded encryption key from: {KEY_FILE if os.path.exists(KEY_FILE) else 'DB_ENCRYPTION_KEY env var'}")

    client = pymongo.MongoClient(MONGO_URI)
    db = client[DB_NAME]
    collection = db[COLLECTION_NAME]

    collection.delete_many({})

    records = []
    for _ in range(20):
        records.append(generate_photo(key))
    for _ in range(15):
        records.append(generate_photo(key, days_ago=random.randint(0, 30)))
    for _ in range(10):
        records.append(generate_photo(key, expiring_next_month=True))
    for _ in range(5):
        records.append(generate_photo(key, days_ago=random.randint(0, 7)))

    result = collection.insert_many(records)
    print(f"Inserted {len(result.inserted_ids)} records (PHI fields encrypted)")
    print(f"Total: {collection.count_documents({})} documents")


if __name__ == "__main__":
    seed_data()