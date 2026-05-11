package remainder

type Reminder struct {
	ID                  int       `json:"id"`
	PatientID           int       `json:"patient_id"`
	QRLinkID            int       `json:"qr_link_id"`
	Interval            int       `json:"interval_seconds"`
	NextApplicationTime time.Time `json:"next_application_time"`
}
