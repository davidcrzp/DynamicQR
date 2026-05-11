package doctor

type Doctor struct {
	ID            int    `json:"id"`
	LicenseNumber string `json:"license_number"`
	FullName      string `json:"full_name"`
	Email         string `json:"email"`
	PasswordHash  string `json:"-"`
}
