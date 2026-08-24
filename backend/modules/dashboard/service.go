package dashboard

import "baseadmin/backend/modules/peserta"

// attendanceLabels mirrors peserta/export.go's attendanceStatusLabels (kept
// unexported there) — a nil/unrecognized status reads the same as
// peserta.StatusBelumKonfirmasi everywhere else in the app.
var attendanceLabels = map[string]string{
	peserta.StatusBelumKonfirmasi:   "Belum Konfirmasi",
	peserta.StatusHadirBelumLengkap: "Hadir, Belum Isi Form",
	peserta.StatusTidakHadir:        "Tidak Hadir",
	peserta.StatusHadirLengkap:      "Hadir, Data Lengkap",
}

// attendanceOrder fixes a stable display order for the attendance chart,
// independent of whatever order SQL GROUP BY happens to return.
var attendanceOrder = []string{
	peserta.StatusBelumKonfirmasi,
	peserta.StatusHadirBelumLengkap,
	peserta.StatusHadirLengkap,
	peserta.StatusTidakHadir,
}

// dietaryOrder mirrors peserta/import.go's importDietaryOptions list — the
// fixed set of options the peserta form/importer offer, plus "Belum Diisi"
// for peserta who haven't set one yet. Each option string doubles as its own
// display label, so no separate label map is needed.
var dietaryOrder = []string{"tidak ada pantangan", "tidak makan daging", "tidak makan ayam", "tidak makan seafood", "vegetarian", "Belum Diisi"}

// Service holds business rules for the dashboard module.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type NamedCount struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Count int64  `json:"count"`
}

type Summary struct {
	TotalPeserta int64        `json:"total_peserta"`
	LoggedIn     int64        `json:"logged_in"`
	NotLoggedIn  int64        `json:"not_logged_in"`
	Attendance   []NamedCount `json:"attendance"`
	Dietary      []NamedCount `json:"dietary"`
	QrGates      []NamedCount `json:"qr_gates"`
}

func (s *Service) Summary() (*Summary, error) {
	total, err := s.repo.TotalPeserta()
	if err != nil {
		return nil, err
	}
	loggedIn, err := s.repo.LoggedInCount()
	if err != nil {
		return nil, err
	}

	attendanceRows, err := s.repo.AttendanceCounts()
	if err != nil {
		return nil, err
	}
	attendanceByKey := make(map[string]int64, len(attendanceRows))
	for _, row := range attendanceRows {
		key := peserta.StatusBelumKonfirmasi
		if row.Value != nil && *row.Value != "" {
			key = *row.Value
		}
		attendanceByKey[key] += row.Count
	}
	attendance := make([]NamedCount, 0, len(attendanceOrder))
	for _, key := range attendanceOrder {
		attendance = append(attendance, NamedCount{Key: key, Label: attendanceLabels[key], Count: attendanceByKey[key]})
	}

	dietaryRows, err := s.repo.DietaryCounts()
	if err != nil {
		return nil, err
	}
	dietaryByKey := make(map[string]int64, len(dietaryRows))
	for _, row := range dietaryRows {
		key := "Belum Diisi"
		if row.Value != nil && *row.Value != "" {
			key = *row.Value
		}
		dietaryByKey[key] += row.Count
	}
	dietary := make([]NamedCount, 0, len(dietaryOrder))
	for _, key := range dietaryOrder {
		dietary = append(dietary, NamedCount{Key: key, Label: key, Count: dietaryByKey[key]})
	}

	gateRows, err := s.repo.GateScanCounts()
	if err != nil {
		return nil, err
	}
	qrGates := make([]NamedCount, 0, len(gateRows))
	for _, row := range gateRows {
		qrGates = append(qrGates, NamedCount{Key: row.GateName, Label: row.GateName, Count: row.Count})
	}

	return &Summary{
		TotalPeserta: total,
		LoggedIn:     loggedIn,
		NotLoggedIn:  total - loggedIn,
		Attendance:   attendance,
		Dietary:      dietary,
		QrGates:      qrGates,
	}, nil
}
