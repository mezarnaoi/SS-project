import random
from datetime import datetime, timedelta
import pymongo

MONGO_URI = "mongodb://admin:supersecret@localhost:27019/"
DB_NAME = "mqtt-streaming-server"
COLLECTION_NAME = "photos"

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

def generate_photo(days_ago=None, expiring_next_month=False):
    now = datetime.now()

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

    # Corectat la UPPERCASE pentru a face match cu frontend-ul/backend-ul
    aviz_types = ["APT", "APT CONDITIONAT", "INAPT TEMPORAR", "INAPT"]
    selected_aviz = random.choices(aviz_types, weights=[70, 15, 10, 5], k=1)[0]

    company = random.choice(COMPANIES)
    unit = random.choice(MEDICAL_UNITS)
    nume = random.choice(SURNAMES)
    prenume = random.choice(NAMES)
    
    # Generare date pentru rapoarte de performanta OCR
    ocr_conf = random.uniform(70.0, 99.9)
    needs_review = ocr_conf < 95.0

    return {
        "timestamp": timestamp,
        "image_type": "jpeg",
        "device_id": f"device-{random.randint(1, 5)}",
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
        "cnp": f"{random.randint(1,2)}{random.randint(50,99)}{random.randint(10,12)}{random.randint(10,28)}123456",
        "profesie_functie": random.choice(JOBS),
        "loc_de_munca": company[0],
        "tip_control": f"Control {selected_control}",
        "aviz_medical": selected_aviz,
        "recomandari": "Nicio recomandare" if selected_aviz == "APT" else "Reevaluare necesara",
        "data": timestamp,
        "data_urm_examinari": data_urm,
        
        # Metrice noi adaugate pentru tab-ul Performance
        "ocr_confidence": ocr_conf,
        "needs_review": needs_review,
        "reviewed_by": "admin" if needs_review and random.choice([True, False]) else None,
        "processing_time_ms": random.randint(200, 1800)
    }

def seed_data():
    try:
        client = pymongo.MongoClient(MONGO_URI)
        db = client[DB_NAME]
        collection = db[COLLECTION_NAME]

        # Curata datele vechi ca sa nu se amestece (optional, dar recomandat pt testare curata)
        collection.delete_many({})

        records = []
        for _ in range(20): records.append(generate_photo())
        for _ in range(15): records.append(generate_photo(days_ago=random.randint(0, 30)))
        for _ in range(10): records.append(generate_photo(expiring_next_month=True))
        for _ in range(5):  records.append(generate_photo(days_ago=random.randint(0, 7)))

        result = collection.insert_many(records)
        print(f"Inserted {len(result.inserted_ids)} records!")
        print(f"Total: {collection.count_documents({})} documents")

    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    seed_data()