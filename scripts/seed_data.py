import random
from datetime import datetime, timedelta
import pymongo

MONGO_URI = "mongodb://admin:supersecret@localhost:27019/"
DB_NAME = "mqtt-streaming-server"
COLLECTION_NAME = "photos"

NAMES = ["Ion", "Maria", "Andrei", "Elena", "Radu", "Ana", "George", "Ioana",
         "Mihai", "Cristina", "Alexandru", "Gabriela", "Florin", "Daniela", "Vlad",
         "Bogdan", "Roxana", "Catalin", "Simona", "Adrian"]
SURNAMES = ["Popescu", "Ionescu", "Dumitru", "Stoica", "Radu", "Gheorghe", "Matei",
            "Florea", "Costea", "Marinescu", "Dinu", "Toma", "Stanciu", "Neagu", "Preda"]
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
    ("Policlinica Medvita", "Str. Lalelelor nr 7", "0700300400"),
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

    aviz_types = ["APT", "APT Conditionat", "Inapt Temporar", "Inapt"]
    selected_aviz = random.choices(aviz_types, weights=[70, 15, 10, 5], k=1)[0]

    company = random.choice(COMPANIES)
    unit = random.choice(MEDICAL_UNITS)
    nume = random.choice(SURNAMES)
    prenume = random.choice(NAMES)

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
        "control_angajare": selected_control == "Angajare",
        "control_periodic": selected_control == "Periodic",
        "control_adaptare": selected_control == "Adaptare",
        "control_reluare": selected_control == "Reluare",
        "control_supraveghere": selected_control == "Supraveghere",
        "control_alte": selected_control == "Alte",
        "aviz_medical": selected_aviz,
        "aviz_apt": selected_aviz == "APT",
        "aviz_apt_conditionat": selected_aviz == "APT Conditionat",
        "aviz_inapt_temporar": selected_aviz == "Inapt Temporar",
        "aviz_inapt": selected_aviz == "Inapt",
        "recomandari": "Nicio recomandare" if selected_aviz == "APT" else "Reevaluare in 30 zile",
        "data": timestamp,
        "data_urm_examinari": data_urm,
    }

def seed_data():
    try:
        client = pymongo.MongoClient(MONGO_URI)
        db = client[DB_NAME]
        collection = db[COLLECTION_NAME]

        records = []

        # 20 random records spread over last year
        for _ in range(20):
            records.append(generate_photo())

        # 15 records from last 30 days (for "last month" reports)
        for i in range(15):
            records.append(generate_photo(days_ago=random.randint(0, 30)))

        # 10 records expiring next month (for expiry reports)
        for _ in range(10):
            records.append(generate_photo(expiring_next_month=True))

        # 5 records from last week
        for _ in range(5):
            records.append(generate_photo(days_ago=random.randint(0, 7)))

        result = collection.insert_many(records)
        print(f"Inserted {len(result.inserted_ids)} records!")
        print(f"Total: {collection.count_documents({})} documents")

    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    seed_data()
