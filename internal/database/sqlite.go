package database

import (
	"fmt"
    "database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteClient struct {
    db *sql.DB
}

func NewDB(path string) (*SQLiteClient, error) {
    db, err := sql.Open("sqlite3", path)
    if err != nil {
        return nil, err
    }
    
	query := `
	PRAGMA foreign_keys = ON;

	CREATE TABLE IF NOT EXISTS doctors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		license_number TEXT UNIQUE,
		full_name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS patients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		doctor_id INTEGER NOT NULL,
		full_name TEXT NOT NULL,
		email TEXT UNIQUE,
		phone_number TEXT,
		FOREIGN KEY (doctor_id) REFERENCES doctors(id) ON DELETE RESTRICT
	);

	CREATE TABLE IF NOT EXISTS treatment_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		patient_id INTEGER NOT NULL,
		qr_link_id INTEGER NOT NULL,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		photo_path TEXT,
		patient_comment TEXT,
		doctor_feedback TEXT,
		is_analyzed INTEGER DEFAULT 0,
		FOREIGN KEY (patient_id) REFERENCES patients(id) ON DELETE RESTRICT,
		FOREIGN KEY (qr_link_id) REFERENCES qr_links(id) ON DELETE SET NULL
	);

	CREATE TABLE IF NOT EXISTS qr_links (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		patient_id INTEGER NOT NULL,
		medication_type TEXT NOT NULL,
		secure_token TEXT UNIQUE NOT NULL,
		is_spent INTEGER DEFAULT 0,
		FOREIGN KEY (patient_id) REFERENCES patients(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS reminders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		patient_id INTEGER NOT NULL,
		qr_link_id INTEGER NOT NULL,
		interval INTEGER NOT NULL,
		next_application_time DATETIME NOT NULL,
		FOREIGN KEY (patient_id) REFERENCES patients(id) ON DELETE CASCADE,
		FOREIGN KEY (qr_link_id) REFERENCES qr_links(id) ON DELETE SET NULL
	);
	`

    _, err = db.Exec(query)
    return &SQLiteClient{db: db}, err
}

func (s *SQLiteClient) Close() error {
    if s.db != nil {
        return s.db.Close()
    }

    return nil
}

func (s *SQLiteClient) RegisterDoctor(licence_number string, full_name string, email string, phone_number string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	query := `
		INSERT INTO doctors(license_number, full_name, email, phone_number)
		VALUES(?, ?, ?, ?);
	`

	_, err = s.db.Exec(query, licence_number, full_name, email, phone_number)
	if err != nil {
		return fmt.Errorf("Could not register doctor: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Commit failed: %v", err)
	}

    return nil
}

func (s *SQLiteClient) RegisterPatient(doctor_id int, full_name string, email string, phone_number string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO patients (doctor_id, full_name, email, phone_number)
		SELECT id, ?, ?, ?
		FROM doctors
		WHERE id = ?;
	`

	res, err := s.db.Exec(query, full_name, email, phone_number, doctor_id) 
	if err != nil {
		return fmt.Errorf("Could not register patient: %v", err)
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("Invalid doctor ID: no record inserted")
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Commit failed: %v", err)
	}

    return nil
}

func (s *SQLiteClient) RegisterTreatment(secureToken string, photoPath, comment string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO treatment_logs (patient_id, qr_link_id, photo_path,	patient_comment)
		SELECT patient_id, id, ?, ?
		FROM qr_links
		WHERE secure_token = ?;
	`

	res, err := tx.Exec(query, photoPath, comment, secureToken)
	if err != nil {
		return fmt.Errorf("Could not register treatment log: %v", err)
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("Invalid token: no record inserted")
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Commit failed: %v", err)
	}

	return nil
}

func (s *SQLiteClient) RegisterQRLink(patient_id int, medication_type string, secure_token string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO qr_links(patient_id, medication_type, secure_token)
		SELECT id, ?, ?
		FROM patients
		WHERE id = ?;
	`

	res, err := s.db.Exec(query, medication_type, secure_token, patient_id)
	if err != nil {
		return fmt.Errorf("Could not register QR: %v", err)
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("Invalid patient ID: no record inserted")
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Commit failed: %v", err)
	}

	return nil
}

func (s *SQLiteClient) RegisterReminder(medication_type string, interval int, secure_token string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO reminders(patient_id, qr_link_id, medication_type, interval)
		SELECT patient_id, id, ?, ?
		FROM qr_links
		WHERE secure_token = ?;
	`

	res, err := s.db.Exec(query, medication_type, interval, secure_token)
	if err != nil {
		return fmt.Errorf("Could not register reminder: %v", err)
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("Invalid patient ID: no record inserted")
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Commit failed: %v", err)
	}

	return nil
}

func (s *SQLiteClient) UpdateReminderDate(secure_token string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE reminders
		SET next_application_time = datetime(next_application_time, '+' || reminders.interval || ' hours')
		FROM qr_links
		WHERE reminders.qr_link_id = qr_links.id
		  AND qr_links.secure_token = ?;
	`

	res, err := s.db.Exec(query, secure_token)
	if err != nil {
		return fmt.Errorf("Could not update reminder: %v", err)
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("Invalid token: no record inserted")
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Commit failed: %v", err)
	}

	return nil
}

// func (s *SQLiteClient) SelectDoctors() ([]Doctor, error) {
// 	tx, err := s.db.Begin()
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer tx.Rollback()
//
// 	query := `
// 		SELECT license_number, full_name, email, password_hash
// 		FROM doctors
// 	`
//
// 	rows, err := tx.Query(query)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
//
// 	var doctors []Doctor
// 	for rows.Next() {
// 		var d Doctor
// 		if err := rows.Scan(&d.LicenseNumber, &d.FullName, &d.Email, &d.PasswordHash); err != nil {
// 			return nil, err
// 		}
// 		doctors = append(doctors, d)
// 	}
//
// 	return doctors, tx.Commit()
// }
//
// // SelectPatients fetches all patient records
// func (s *SQLiteClient) SelectPatients() ([]Patient, error) {
// 	tx, err := s.DB.Begin()
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer tx.Rollback()
//
// 	rows, err := tx.Query("SELECT id, doctor_id, full_name, email, phone_number FROM patients")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
//
// 	var patients []Patient
// 	for rows.Next() {
// 		var p Patient
// 		if err := rows.Scan(&p.ID, &p.DoctorID, &p.FullName, &p.Email, &p.PhoneNumber); err != nil {
// 			return nil, err
// 		}
// 		patients = append(patients, p)
// 	}
//
// 	return patients, tx.Commit()
// }
//
// // SelectTreatmentLogs fetches all treatment logs
// func (s *SQLiteClient) SelectTreatmentLogs() ([]TreatmentLog, error) {
// 	tx, err := s.DB.Begin()
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer tx.Rollback()
//
// 	rows, err := tx.Query("SELECT id, patient_id, qr_link_id, applied_at, photo_path, patient_comment, doctor_feedback, is_analyzed FROM treatment_logs")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
//
// 	var logs []TreatmentLog
// 	for rows.Next() {
// 		var l TreatmentLog
// 		if err := rows.Scan(&l.ID, &l.PatientID, &l.QRLinkID, &l.AppliedAt, &l.PhotoPath, &l.PatientComment, &l.DoctorFeedback, &l.IsAnalyzed); err != nil {
// 			return nil, err
// 		}
// 		logs = append(logs, l)
// 	}
//
// 	return logs, tx.Commit()
// }
//
// // SelectQR fetches all QR link records
// func (s *SQLiteClient) SelectQR() ([]QRLink, error) {
// 	tx, err := s.DB.Begin()
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer tx.Rollback()
//
// 	rows, err := tx.Query("SELECT id, patient_id, medication_type, secure_token, is_spent FROM qr_links")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
//
// 	var qrs []QRLink
// 	for rows.Next() {
// 		var q QRLink
// 		if err := rows.Scan(&q.ID, &q.PatientID, &q.MedicationType, &q.Secure_token, &q.IsSpent); err != nil {
// 			return nil, err
// 		}
// 		qrs = append(qrs, q)
// 	}
//
// 	return qrs, tx.Commit()
// }
//
// // SelectReminders fetches all reminder records
// func (s *SQLiteClient) SelectReminders() ([]Reminder, error) {
// 	tx, err := s.DB.Begin()
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer tx.Rollback()
//
// 	rows, err := tx.Query("SELECT id, patient_id, qr_link_id, interval, next_application_time, is_sent FROM reminders")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
//
// 	var reminders []Reminder
// 	for rows.Next() {
// 		var r Reminder
// 		if err := rows.Scan(&r.ID, &r.PatientID, &r.QRLinkID, &r.Interval, &r.NextApplicationTime, &r.IsSent); err != nil {
// 			return nil, err
// 		}
// 		reminders = append(reminders, r)
// 	}
//
// 	return reminders, tx.Commit()
// }

// func (s *SQLiteClient) RegisterRemainder() {}
// func (s *SQLiteClient) GetTodos() ([]domain.Todo, error) {
//     rows, _ := s.db.Query("SELECT id, task, done FROM todos")
//     var todos []domain.Todo
//     for rows.Next() {
//         var t domain.Todo
//         rows.Scan(&t.ID, &t.Task, &t.Done)
//         todos = append(todos, t)
//     }
//
//     return todos, nil
// }
//
// func (s *SQLiteClient) PostTodo(task string, done bool) error {
//     _, err = s.db.Exec("INSERT INTO todos(task, done) VALUES(?, ?)", task, done)
// 	if err != nil {
// 		return fmt.Errorf("could not insert user: %v", err)
// 	}
//     return nil
// }
