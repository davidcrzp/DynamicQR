package treatment

import "time"

type TreatmentLog struct {
    ID             int64     `json:"id"`
    PatientID      int64     `json:"patient_id"`
    AppliedAt      time.Time `json:"applied_at"`
    PhotoURL       string    `json:"photo_url"`
    PatientComment string    `json:"patient_comment"`
    DoctorFeedback string    `json:"doctor_feedback"` 
    IsAnalyzed     bool      `json:"is_analyzed"`
}

type QRLink struct {
    ID             int64  `json:"id"`
    PatientID      int64  `json:"patient_id"`
    SecureToken    string `json:"secure_token"`
    MedicationType string `json:"medication_type"`
    IsActive       bool   `json:"is_active"`
}
