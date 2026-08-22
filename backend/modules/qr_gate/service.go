package qr_gate

import (
	"errors"
	"time"

	"baseadmin/backend/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotFound       = errors.New("qr gate not found")
	ErrCodeNotFound   = errors.New("qr code not found")
	ErrAlreadyClaimed = errors.New("qr gate already claimed by another participant")
)

// Service holds business rules for the qr_gate module.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(p utils.Pagination) ([]QrGateWithStats, int64, error) {
	return s.repo.List(p)
}

func (s *Service) Get(uuidStr string) (*QrGate, error) {
	gate, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, ErrNotFound
	}
	return gate, nil
}

// ScanByUser is the shape returned for the admin "who scanned this gate" view.
type ScanByUser struct {
	UUID          uuid.UUID `json:"uuid"`
	PointsAwarded int       `json:"points_awarded"`
	ScannedAt     time.Time `json:"scanned_at"`
	UserName      string    `json:"user_name"`
	KtpNumber     string    `json:"ktp_number"`
}

func (s *Service) GateScans(uuidStr string) ([]ScanByUser, error) {
	gate, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, ErrNotFound
	}
	rows, err := s.repo.ListScansForGate(gate.ID)
	if err != nil {
		return nil, err
	}
	result := make([]ScanByUser, 0, len(rows))
	for _, row := range rows {
		ktp := ""
		if row.User.KtpNumber != nil {
			ktp = *row.User.KtpNumber
		}
		result = append(result, ScanByUser{
			UUID: row.UUID, PointsAwarded: row.PointsAwarded, ScannedAt: row.CreatedAt,
			UserName: row.User.Name, KtpNumber: ktp,
		})
	}
	return result, nil
}

type Input struct {
	Name       string
	Code       string
	Points     int
	IsReusable bool
	ActorID    *uint64
}

func (s *Service) Create(in Input) (*QrGate, error) {
	gate := QrGate{Name: in.Name, Code: in.Code, Points: in.Points, IsReusable: in.IsReusable}
	gate.CreatedBy = in.ActorID
	if err := s.repo.Create(&gate); err != nil {
		return nil, err
	}
	return &gate, nil
}

func (s *Service) Update(uuidStr string, in Input) (before *QrGate, after *QrGate, err error) {
	gate, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, nil, ErrNotFound
	}
	old := *gate

	gate.Name = in.Name
	gate.Code = in.Code
	gate.Points = in.Points
	gate.IsReusable = in.IsReusable
	gate.UpdatedBy = in.ActorID

	if err := s.repo.Save(gate); err != nil {
		return nil, nil, err
	}
	return &old, gate, nil
}

func (s *Service) Delete(uuidStr string, actorID *uint64) (*QrGate, error) {
	gate, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, ErrNotFound
	}
	if err := s.repo.Delete(gate, actorID); err != nil {
		return nil, err
	}
	return gate, nil
}

// ScanResult is what Scan returns on every non-error outcome. AlreadyScanned
// distinguishes "you already claimed this gate's points" from a fresh
// redemption — it's not treated as a failure (the participant did scan a
// valid code, they just don't get points twice), so callers should render
// it as an informational success, not an error.
type ScanResult struct {
	AlreadyScanned bool
	Gate           QrGate
	PointsAwarded  int
	ScanUUID       uuid.UUID
	ScannedAt      time.Time
}

// Scan records a participant's redemption of a QR gate's code. Everything
// runs inside one transaction with the gate row locked (see
// Repository.FindByCodeTx) so two simultaneous scans of a one-time-only gate
// can't both slip past the "already claimed" check.
func (s *Service) Scan(userID uint64, code string) (*ScanResult, error) {
	var result ScanResult

	err := s.repo.DB.Transaction(func(tx *gorm.DB) error {
		gate, err := s.repo.FindByCodeTx(tx, code)
		if err != nil {
			return ErrCodeNotFound
		}
		result.Gate = *gate

		alreadyScanned, err := s.repo.HasUserScannedTx(tx, gate.ID, userID)
		if err != nil {
			return err
		}
		if alreadyScanned {
			result.AlreadyScanned = true
			return nil
		}

		if !gate.IsReusable {
			count, err := s.repo.CountScansForGateTx(tx, gate.ID)
			if err != nil {
				return err
			}
			if count > 0 {
				return ErrAlreadyClaimed
			}
		}

		scan := QrGateScan{QrGateID: gate.ID, UserID: userID, PointsAwarded: gate.Points}
		if err := s.repo.CreateScanTx(tx, &scan); err != nil {
			return err
		}
		result.PointsAwarded = scan.PointsAwarded
		result.ScanUUID = scan.UUID
		result.ScannedAt = scan.CreatedAt
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ScanWithGate is the shape returned for a participant's own scan history.
type ScanWithGate struct {
	UUID          uuid.UUID `json:"uuid"`
	PointsAwarded int       `json:"points_awarded"`
	ScannedAt     time.Time `json:"scanned_at"`
	GateName      string    `json:"gate_name"`
}

func (s *Service) MyScans(userID uint64) (scans []ScanWithGate, totalPoints int, err error) {
	rows, err := s.repo.ListScansForUser(userID)
	if err != nil {
		return nil, 0, err
	}
	scans = make([]ScanWithGate, 0, len(rows))
	for _, row := range rows {
		totalPoints += row.PointsAwarded
		scans = append(scans, ScanWithGate{
			UUID: row.UUID, PointsAwarded: row.PointsAwarded, ScannedAt: row.CreatedAt, GateName: row.QrGate.Name,
		})
	}
	return scans, totalPoints, nil
}
