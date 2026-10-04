package qr_gate

import (
	"errors"
	"sort"
	"time"

	"baseadmin/backend/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotFound       = errors.New("qr gate not found")
	ErrCodeNotFound   = errors.New("qr code not found")
	ErrAlreadyClaimed = errors.New("qr gate already claimed by another participant")
	ErrManualNotFound = errors.New("manual point not found")
	ErrScanNotFound   = errors.New("scan not found")
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
	return s.redeem(userID, func(tx *gorm.DB) (*QrGate, error) { return s.repo.FindByCodeTx(tx, code) })
}

// Checkin is the admin-triggered equivalent of Scan — same rules and locking,
// but the gate is picked by UUID from the admin panel instead of decoded from
// a camera scan (e.g. a participant whose phone can't scan at the booth).
func (s *Service) Checkin(userID uint64, gateUUID string) (*ScanResult, error) {
	return s.redeem(userID, func(tx *gorm.DB) (*QrGate, error) { return s.repo.FindByUUIDTx(tx, gateUUID) })
}

func (s *Service) redeem(userID uint64, findGate func(tx *gorm.DB) (*QrGate, error)) (*ScanResult, error) {
	var result ScanResult

	err := s.repo.DB.Transaction(func(tx *gorm.DB) error {
		gate, err := findGate(tx)
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

// GateCheckinStatus is one QR gate as seen from a single participant's point
// history — drives the admin's manual check-in buttons. ClaimedByOther marks
// a one-time gate someone else already took, so it can't be checked in.
type GateCheckinStatus struct {
	UUID           uuid.UUID `json:"uuid"`
	Name           string    `json:"name"`
	Points         int       `json:"points"`
	IsReusable     bool      `json:"is_reusable"`
	Scanned        bool      `json:"scanned"`
	ClaimedByOther bool      `json:"claimed_by_other"`
}

func (s *Service) GateStatusForUser(userID uint64) ([]GateCheckinStatus, error) {
	rows, err := s.repo.GateStatusForUser(userID)
	if err != nil {
		return nil, err
	}
	list := make([]GateCheckinStatus, 0, len(rows))
	for _, row := range rows {
		list = append(list, GateCheckinStatus{
			UUID: row.UUID, Name: row.Name, Points: row.Points, IsReusable: row.IsReusable, Scanned: row.Scanned,
			ClaimedByOther: !row.IsReusable && !row.Scanned && row.ScanCount > 0,
		})
	}
	return list, nil
}

// Point history sources — "manual" entries are admin-granted ManualPoints,
// merged into the same list as QR scans.
const (
	SourceScan   = "scan"
	SourceManual = "manual"
)

// ScanWithGate is one entry in a participant's point history. GateName holds
// the ManualPoint's name for manual entries, so the website's history list
// renders both kinds without special-casing.
type ScanWithGate struct {
	UUID          uuid.UUID `json:"uuid"`
	PointsAwarded int       `json:"points_awarded"`
	ScannedAt     time.Time `json:"scanned_at"`
	GateName      string    `json:"gate_name"`
	Source        string    `json:"source"`
}

// LeaderboardEntry is one ranked participant as shown on the website and the
// public monitor display — Rank is computed in SQL (see
// rankedLeaderboardSQL), not stored.
type LeaderboardEntry struct {
	Rank   int    `json:"rank"`
	Name   string `json:"name"`
	Points int64  `json:"points"`
}

// MyLeaderboard is the logged-in participant's view: the top list plus their
// own position, so the website can pin them below the list when they rank
// outside it. Me is nil for accounts that aren't on the board.
type MyLeaderboard struct {
	Entries []LeaderboardEntry `json:"entries"`
	Me      *LeaderboardEntry  `json:"me"`
}

// PublicLeaderboardEntry is what the login-free monitor display shows — the
// NIP is included on purpose so attendees can tell same-name participants
// apart on screen.
type PublicLeaderboardEntry struct {
	LeaderboardEntry
	KtpNumber string `json:"ktp_number"`
}

// AdminLeaderboardEntry adds the identifying/stat columns only the admin
// panel should see.
type AdminLeaderboardEntry struct {
	LeaderboardEntry
	UUID      uuid.UUID `json:"uuid"`
	KtpNumber string    `json:"ktp_number"`
	ScanCount int64     `json:"scan_count"`
}

// clampLeaderboardLimit keeps an unbounded/absurd query param from forcing a
// full-board payload on the participant/public endpoints.
func clampLeaderboardLimit(limit int) int {
	if limit <= 0 || limit > 100 {
		return 10
	}
	return limit
}

func toEntry(row LeaderboardRow) LeaderboardEntry {
	return LeaderboardEntry{Rank: row.RankNo, Name: row.Name, Points: row.Points}
}

func ktpOf(row LeaderboardRow) string {
	if row.KtpNumber == nil {
		return ""
	}
	return *row.KtpNumber
}

func (s *Service) PublicLeaderboard(limit int) ([]PublicLeaderboardEntry, error) {
	rows, err := s.repo.Leaderboard(clampLeaderboardLimit(limit))
	if err != nil {
		return nil, err
	}
	list := make([]PublicLeaderboardEntry, 0, len(rows))
	for _, row := range rows {
		list = append(list, PublicLeaderboardEntry{LeaderboardEntry: toEntry(row), KtpNumber: ktpOf(row)})
	}
	return list, nil
}

func (s *Service) Leaderboard(limit int) ([]LeaderboardEntry, error) {
	rows, err := s.repo.Leaderboard(clampLeaderboardLimit(limit))
	if err != nil {
		return nil, err
	}
	list := make([]LeaderboardEntry, 0, len(rows))
	for _, row := range rows {
		list = append(list, toEntry(row))
	}
	return list, nil
}

func (s *Service) MyLeaderboard(limit int, userID uint64) (*MyLeaderboard, error) {
	entries, err := s.Leaderboard(limit)
	if err != nil {
		return nil, err
	}
	result := &MyLeaderboard{Entries: entries}
	pos, err := s.repo.LeaderboardPosition(userID)
	if err != nil {
		return nil, err
	}
	if pos != nil {
		me := toEntry(*pos)
		result.Me = &me
	}
	return result, nil
}

func (s *Service) AdminLeaderboard() ([]AdminLeaderboardEntry, error) {
	rows, err := s.repo.Leaderboard(0)
	if err != nil {
		return nil, err
	}
	list := make([]AdminLeaderboardEntry, 0, len(rows))
	for _, row := range rows {
		list = append(list, AdminLeaderboardEntry{LeaderboardEntry: toEntry(row), UUID: row.UUID, KtpNumber: ktpOf(row), ScanCount: row.ScanCount})
	}
	return list, nil
}

func (s *Service) MyScans(userID uint64) (scans []ScanWithGate, totalPoints int, err error) {
	rows, err := s.repo.ListScansForUser(userID)
	if err != nil {
		return nil, 0, err
	}
	manual, err := s.repo.ListManualPointsForUser(userID)
	if err != nil {
		return nil, 0, err
	}
	scans = make([]ScanWithGate, 0, len(rows)+len(manual))
	for _, row := range rows {
		totalPoints += row.PointsAwarded
		scans = append(scans, ScanWithGate{
			UUID: row.UUID, PointsAwarded: row.PointsAwarded, ScannedAt: row.CreatedAt, GateName: row.QrGate.Name, Source: SourceScan,
		})
	}
	for _, mp := range manual {
		totalPoints += mp.Points
		scans = append(scans, ScanWithGate{
			UUID: mp.UUID, PointsAwarded: mp.Points, ScannedAt: mp.CreatedAt, GateName: mp.Name, Source: SourceManual,
		})
	}
	sort.SliceStable(scans, func(i, j int) bool { return scans[i].ScannedAt.After(scans[j].ScannedAt) })
	return scans, totalPoints, nil
}

func (s *Service) AddManualPoint(userID uint64, name string, points int, actorID *uint64) (*ManualPoint, error) {
	mp := ManualPoint{UserID: userID, Name: name, Points: points}
	mp.CreatedBy = actorID
	if err := s.repo.CreateManualPoint(&mp); err != nil {
		return nil, err
	}
	return &mp, nil
}

func (s *Service) DeleteManualPoint(userID uint64, uuidStr string, actorID *uint64) (*ManualPoint, error) {
	mp, err := s.repo.FindManualPoint(userID, uuidStr)
	if err != nil {
		return nil, ErrManualNotFound
	}
	if err := s.repo.DeleteManualPoint(mp, actorID); err != nil {
		return nil, err
	}
	return mp, nil
}

// DeleteScan undoes a check-in/scan (admin correction), freeing the gate so
// the participant can check in to it again — and, for a one-time gate, so
// another participant can claim it.
func (s *Service) DeleteScan(userID uint64, uuidStr string) (*QrGateScan, error) {
	scan, err := s.repo.FindScanForUser(userID, uuidStr)
	if err != nil {
		return nil, ErrScanNotFound
	}
	if err := s.repo.DeleteScan(scan); err != nil {
		return nil, err
	}
	return scan, nil
}
