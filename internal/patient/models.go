package patient

import "time"

type Patient struct {
    ID        int64     `json:"id"`
    FullName  string    `json:"full_name"`
    Email     string    `json:"email"`
    Phone     string    `json:"phone"`
    CreatedAt time.Time `json:"created_at"`
}
