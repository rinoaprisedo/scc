package dashboard

import (
	"time"

	"baseadmin/backend/modules/peserta"
	"baseadmin/backend/modules/qris_cross_border"
)

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

// qrisStatusLabels mirrors the frontend's own statusLabel map
// (pages/qris_cross_border/QrisCrossBorder.jsx) — same precedent as
// attendanceLabels above.
var qrisStatusLabels = map[string]string{
	qris_cross_border.StatusPending:         "Pending (Diproses AI)",
	qris_cross_border.StatusWaitingApproval: "Waiting Approval",
	qris_cross_border.StatusApproved:        "Approved",
	qris_cross_border.StatusRejected:        "Rejected",
}

var qrisStatusOrder = []string{
	qris_cross_border.StatusPending,
	qris_cross_border.StatusWaitingApproval,
	qris_cross_border.StatusApproved,
	qris_cross_border.StatusRejected,
}

// QrisCrossBorderTop is one row of the dashboard's "highest nominal" table —
// flattened the same way qris_cross_border.Response is, instead of
// serializing the full users.User relation.
type QrisCrossBorderTop struct {
	PesertaName   string    `json:"peserta_name"`
	PesertaKtp    string    `json:"peserta_ktp_number"`
	MerchantName  string    `json:"merchant_name"`
	NominalAsing  float64   `json:"nominal_asing"`
	NominalRupiah float64   `json:"nominal_rupiah"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// qrisCrossBorderTopLimit caps the dashboard's "highest nominal" table —
// a fixed top-N list, not a paginated one.
const qrisCrossBorderTopLimit = 20

// Service holds business rules for the dashboard module.
type Service struct {
	repo    *Repository
	peserta *peserta.Service
}

func NewService(repo *Repository, pesertaService *peserta.Service) *Service {
	return &Service{repo: repo, peserta: pesertaService}
}

type NamedCount struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Count int64  `json:"count"`
}

// BlazerSizeStock is one blazer size's label, its admin-set initial quota
// (Stock), how many peserta already have it assigned, and the computed
// Remaining (Stock - Assigned) — Stock alone is never a live remaining
// count, see BlazerSizeStockRow's comment in repository.go.
type BlazerSizeStock struct {
	Size      string `json:"size"`
	Stock     int    `json:"stock"`
	Assigned  int64  `json:"assigned"`
	Remaining int64  `json:"remaining"`
}

type Summary struct {
	TotalPeserta       int64                `json:"total_peserta"`
	LoggedIn           int64                `json:"logged_in"`
	NotLoggedIn        int64                `json:"not_logged_in"`
	Attendance         []NamedCount         `json:"attendance"`
	Dietary            []NamedCount         `json:"dietary"`
	QrGates            []NamedCount         `json:"qr_gates"`
	BlazerSizes        []BlazerSizeStock    `json:"blazer_sizes"`
	QrisCrossBorder    []NamedCount         `json:"qris_cross_border_status"`
	QrisCrossBorderTop []QrisCrossBorderTop `json:"qris_cross_border_top"`
}

func (s *Service) Summary() (*Summary, error) {
	// Regenerated on every dashboard load rather than on a write path or a
	// schedule — piggybacks here so a stale attendance_status (data complete
	// but status not, or vice versa) is caught the moment an admin looks at
	// the dashboard, with no separate job to wire up. See
	// peserta.Service.RegenerateAttendanceStatuses for why some statuses are
	// left untouched.
	if _, err := s.peserta.RegenerateAttendanceStatuses(); err != nil {
		return nil, err
	}

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

	blazerRows, err := s.repo.BlazerSizeStockRows()
	if err != nil {
		return nil, err
	}
	blazerSizes := make([]BlazerSizeStock, 0, len(blazerRows))
	for _, row := range blazerRows {
		blazerSizes = append(blazerSizes, BlazerSizeStock{
			Size:      row.Size,
			Stock:     row.Stock,
			Assigned:  row.Assigned,
			Remaining: int64(row.Stock) - row.Assigned,
		})
	}

	qrisStatusRows, err := s.repo.QrisCrossBorderStatusCounts()
	if err != nil {
		return nil, err
	}
	qrisStatusByKey := make(map[string]int64, len(qrisStatusRows))
	for _, row := range qrisStatusRows {
		if row.Value != nil {
			qrisStatusByKey[*row.Value] += row.Count
		}
	}
	qrisCrossBorder := make([]NamedCount, 0, len(qrisStatusOrder))
	for _, key := range qrisStatusOrder {
		qrisCrossBorder = append(qrisCrossBorder, NamedCount{Key: key, Label: qrisStatusLabels[key], Count: qrisStatusByKey[key]})
	}

	qrisTopRows, err := s.repo.QrisCrossBorderTopNominal(qrisCrossBorderTopLimit, qris_cross_border.StatusApproved)
	if err != nil {
		return nil, err
	}
	qrisCrossBorderTop := make([]QrisCrossBorderTop, 0, len(qrisTopRows))
	for _, row := range qrisTopRows {
		merchant, ktp := "", ""
		if row.MerchantName != nil {
			merchant = *row.MerchantName
		}
		if row.PesertaKtp != nil {
			ktp = *row.PesertaKtp
		}
		qrisCrossBorderTop = append(qrisCrossBorderTop, QrisCrossBorderTop{
			PesertaName:   row.PesertaName,
			PesertaKtp:    ktp,
			MerchantName:  merchant,
			NominalAsing:  row.NominalAsing,
			NominalRupiah: row.NominalRupiah,
			Status:        row.Status,
			CreatedAt:     row.CreatedAt,
		})
	}

	return &Summary{
		TotalPeserta:       total,
		LoggedIn:           loggedIn,
		NotLoggedIn:        total - loggedIn,
		Attendance:         attendance,
		Dietary:            dietary,
		QrGates:            qrGates,
		BlazerSizes:        blazerSizes,
		QrisCrossBorder:    qrisCrossBorder,
		QrisCrossBorderTop: qrisCrossBorderTop,
	}, nil
}
