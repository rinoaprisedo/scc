package peserta

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"baseadmin/backend/modules/blazer_sizes"
	"baseadmin/backend/modules/users"
	"baseadmin/backend/session"
	"baseadmin/backend/storage"
	"baseadmin/backend/utils"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const dateLayout = "2006-01-02"

var (
	ErrNotFound                = errors.New("peserta not found")
	ErrDuplicateEntry          = errors.New("failed to create peserta (email or NIK may already be in use)")
	ErrPassportExpiryTooSoon   = errors.New("masa berlaku paspor minimal 6 bulan setelah 10 Oktober 2026 (hingga 10 April 2027)")
	ErrInvalidAttendanceStatus = errors.New("invalid attendance status")
	ErrBlazerSizeInvalid       = errors.New("ukuran blazer tidak ditemukan")
	ErrBlazerSizeOutOfStock    = errors.New("stok ukuran blazer sudah habis")
)

// minPassportExpiry is the event date (2026-10-10) plus the standard
// 6-month passport-validity travel requirement. Only enforced on the
// participant's own self-service save (UpdateProfile) — admin Create/Update
// stay unrestricted since an admin may need to record/correct data as-is.
var minPassportExpiry = time.Date(2027, 4, 10, 0, 0, 0, 0, time.UTC)

// Attendance status — the single field the admin panel and website both
// read to know where a peserta stands, replacing what used to be a
// separate attendance-answer boolean plus an on-the-fly completeness check
// derived from several other fields on every read.
const (
	StatusBelumKonfirmasi   = "belum_konfirmasi"    // hasn't answered the attendance prompt yet
	StatusHadirBelumLengkap = "hadir_belum_lengkap" // answered "hadir", Form/Shirt Size not done
	StatusTidakHadir        = "tidak_hadir"         // answered "tidak hadir" — terminal, ignores form state
	StatusHadirLengkap      = "hadir_lengkap"       // answered "hadir" and Form + Shirt Size are both complete
)

// isValidAttendanceStatus backs the admin panel's manual Kehadiran override
// on Update (see UpdateInput.AttendanceStatus) — guards against an
// arbitrary string being written to a column the frontend otherwise only
// ever sets via a fixed dropdown.
func isValidAttendanceStatus(s string) bool {
	switch s {
	case StatusBelumKonfirmasi, StatusHadirBelumLengkap, StatusTidakHadir, StatusHadirLengkap:
		return true
	default:
		return false
	}
}

func strSet(s *string) bool {
	return s != nil && *s != ""
}

// isFormComplete/isShirtComplete mirror website/src/utils/onboarding.js
// field-for-field. They check the scalar *ID columns (OriginCityID,
// NearestAirportID), not the preloaded OriginCity/NearestAirport structs —
// those get deliberately nilled out alongside the ID on every profile save
// (see the "stale preloaded association" comments below), so checking the
// struct pointer here would read as incomplete right after a save that just
// set a real city/airport.
// MiddleName/LastName are deliberately not required here — plenty of
// participants (esp. Excel-imported ones) genuinely have only a first name,
// and requiring either field left otherwise-fully-filled profiles stuck
// showing as incomplete.
func isFormComplete(u *users.User) bool {
	cityOK := u.OriginCityID != nil || strSet(u.OriginCityOther)
	return strSet(u.Title) &&
		strSet(u.FirstName) &&
		u.BirthDate != nil &&
		cityOK &&
		u.NearestAirportID != nil &&
		strSet(u.DietaryRestriction) &&
		strSet(u.PhoneNumber) &&
		strSet(u.KtpNumber) &&
		strSet(u.KtpFile) &&
		strSet(u.PassportNumber) &&
		strSet(u.PassportFile) &&
		u.PassportExpiry != nil
}

func isShirtComplete(u *users.User) bool {
	return strSet(u.BlazerSize)
}

// recomputeAttendanceStatus keeps AttendanceStatus in sync after any
// profile/shirt-size field changes. "belum_konfirmasi" and "tidak_hadir" are
// left alone — completing profile fields is irrelevant until the
// participant has actually said they're attending. nil (non-Peserta users,
// or defensively any peserta somehow missing a status) is left untouched too.
func recomputeAttendanceStatus(u *users.User) {
	if u.AttendanceStatus == nil {
		return
	}
	switch *u.AttendanceStatus {
	case StatusBelumKonfirmasi, StatusTidakHadir:
		return
	default:
		next := StatusHadirBelumLengkap
		if isFormComplete(u) && isShirtComplete(u) {
			next = StatusHadirLengkap
		}
		u.AttendanceStatus = &next
	}
}

// RegenerateAttendanceStatuses recomputes every peserta's AttendanceStatus
// from their current profile/shirt-size data, fixing rows where the two have
// drifted apart (e.g. a direct DB edit, or any write path added later that
// forgets to call recomputeAttendanceStatus). "belum_konfirmasi" and
// "tidak_hadir" rows are left alone, same as a single-row recompute — see
// recomputeAttendanceStatus. Called on every dashboard load
// (dashboard.Service.Summary) and every Peserta admin list load (List
// below), rather than on a schedule, since this data doesn't change fast
// enough to need one. Returns how many rows were actually changed.
func (s *Service) RegenerateAttendanceStatuses() (int, error) {
	list, err := s.repo.ListAll("")
	if err != nil {
		return 0, err
	}

	updated := 0
	for i := range list {
		user := list[i]
		if user.AttendanceStatus == nil {
			continue
		}
		before := *user.AttendanceStatus
		recomputeAttendanceStatus(&user)
		if *user.AttendanceStatus == before {
			continue
		}
		if err := s.repo.UpdateAttendanceStatus(user.ID, *user.AttendanceStatus); err != nil {
			return updated, err
		}
		updated++
	}
	return updated, nil
}

// placeholderEmailDomain backs the synthetic email generated when a peserta
// is created without one — NIK, not email, is the required identifier here
// (peserta accounts will eventually log in with NIK + password), but the
// shared users table still has a NOT NULL/unique email column used by the
// regular Users module.
const placeholderEmailDomain = "@peserta.local"

// Service holds business rules for the peserta module: password hashing,
// forcing the Peserta role, resolving origin-city/airport relations.
// Repository stays a thin data-access layer, matching the users module.
type Service struct {
	repo        *Repository
	storage     storage.StorageInterface
	redis       *redis.Client
	blazerSizes *blazer_sizes.Repository
}

func NewService(repo *Repository, s storage.StorageInterface, rdb *redis.Client, blazerSizes *blazer_sizes.Repository) *Service {
	return &Service{repo: repo, storage: s, redis: rdb, blazerSizes: blazerSizes}
}

// checkBlazerSizeStock enforces server-side the same stock-availability
// rule the website's size picker only enforces visually (disabling
// out-of-stock <option>s) — admin CRUD must not be able to bypass that by
// selecting/typing an out-of-stock size directly. A size left unchanged is
// always allowed through, even if its stock has since run out, so editing an
// unrelated field on an already-assigned peserta is never blocked by their
// own existing size.
//
// blazer_sizes.Stock is an admin-set initial quota, not a live remaining
// counter — it's never decremented anywhere in the app. Remaining
// availability is always Stock minus how many peserta already have that
// size (see peserta.Repository.CountByBlazerSize and
// dashboard.Repository.BlazerSizeStockRows, which compute it the same way).
func (s *Service) checkBlazerSizeStock(newSize string, currentSize *string) error {
	if newSize == "" {
		return nil
	}
	if currentSize != nil && *currentSize == newSize {
		return nil
	}
	bs, err := s.blazerSizes.FindBySize(newSize)
	if err != nil {
		return ErrBlazerSizeInvalid
	}
	assigned, err := s.repo.CountByBlazerSize(newSize)
	if err != nil {
		return err
	}
	if int64(bs.Stock)-assigned <= 0 {
		return ErrBlazerSizeOutOfStock
	}
	return nil
}

// invalidateSessions mirrors users.Service.invalidateSessions — kept for
// parity in case a Peserta account is ever used to log in and its status
// changes.
func (s *Service) invalidateSessions(userID uint64) {
	_ = session.DeleteAllForUser(context.Background(), s.redis, userID)
	_ = s.repo.DeleteAllSessions(userID)
}

// List regenerates every peserta's AttendanceStatus before reading the page
// back, same as dashboard.Service.Summary — so opening/paging through the
// Peserta admin list also self-heals any row where AttendanceStatus drifted
// out of sync with the underlying profile data.
func (s *Service) List(p utils.Pagination) ([]UserWithPoints, int64, error) {
	if _, err := s.RegenerateAttendanceStatuses(); err != nil {
		return nil, 0, err
	}
	return s.repo.List(p)
}

// ListAll backs CSV/Excel export — every matching row, not one page.
func (s *Service) ListAll(search string) ([]users.User, error) {
	return s.repo.ListAll(search)
}

// OpenKtpFile backs the KTP ZIP export — reads a single uploaded KTP file's
// contents by its stored relative path (users.User.KtpFile).
func (s *Service) OpenKtpFile(path string) (io.ReadCloser, error) {
	return s.storage.Open(path)
}

// OpenPassportFile mirrors OpenKtpFile for the passport ZIP export, reading
// by users.User.PassportFile's stored relative path.
func (s *Service) OpenPassportFile(path string) (io.ReadCloser, error) {
	return s.storage.Open(path)
}

// FileURL resolves a stored relative path (users.User.KtpFile/PassportFile)
// into the full URL the storage backend serves it from — used by the
// CSV/Excel export so those columns hold a directly-openable link rather
// than the bare relative path. Empty when nothing was uploaded.
func (s *Service) FileURL(path *string) string {
	if path == nil || *path == "" {
		return ""
	}
	return s.storage.GetURL(*path)
}

func (s *Service) Get(uuidStr string) (*users.User, error) {
	user, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, ErrNotFound
	}
	return user, nil
}

// ProfileInput holds every Peserta-specific field as plain strings — dates
// as "2006-01-02", UUIDs for relations — so the handler stays a thin
// JSON-binding layer and all parsing/resolution lives here.
type ProfileInput struct {
	Title              string
	FirstName          string
	MiddleName         string
	LastName           string
	BirthDate          string
	OriginCityUUID     string
	OriginCityOther    string
	NearestAirportUUID string
	DietaryRestriction string
	PhoneNumber        string
	Region             string
	Cabang             string
	Position           string
	KtpNumber          string
	NomorKtp           string
	PassportNumber     string
	PassportExpiry     string
	BlazerSize         string
	NomorMeja          string
	Description        string
}

func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return nil
	}
	return &t
}

func ptrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

type CreateInput struct {
	Name     string
	Email    string
	Password string
	Status   string
	Profile  ProfileInput
	ActorID  *uint64
}

func (s *Service) Create(in CreateInput) (*users.User, error) {
	if err := s.checkBlazerSizeStock(in.Profile.BlazerSize, nil); err != nil {
		return nil, err
	}

	// Lowercased before hashing so PesertaLogin's case-insensitive check
	// (website login requirement — NIK/password login is not treated as
	// case-sensitive there) has a matching hash to compare against.
	hashed, err := bcrypt.GenerateFromPassword([]byte(strings.ToLower(in.Password)), 12)
	if err != nil {
		return nil, err
	}

	roleID, err := s.repo.PesertaRoleID()
	if err != nil {
		return nil, err
	}

	cityID, err := s.repo.ResolveCityID(in.Profile.OriginCityUUID)
	if err != nil {
		return nil, err
	}
	airportID, err := s.repo.ResolveAirportID(in.Profile.NearestAirportUUID)
	if err != nil {
		return nil, err
	}

	status := users.StatusActive
	if in.Status != "" {
		status = users.UserStatus(in.Status)
	}

	email := in.Email
	if email == "" {
		email = in.Profile.KtpNumber + placeholderEmailDomain
	}

	attendanceStatus := StatusBelumKonfirmasi

	user := users.User{
		Name:               in.Name,
		Email:              email,
		Password:           string(hashed),
		Status:             status,
		RoleID:             &roleID,
		AttendanceStatus:   &attendanceStatus,
		Title:              ptrOrNil(in.Profile.Title),
		FirstName:          ptrOrNil(in.Profile.FirstName),
		MiddleName:         ptrOrNil(in.Profile.MiddleName),
		LastName:           ptrOrNil(in.Profile.LastName),
		BirthDate:          parseDate(in.Profile.BirthDate),
		OriginCityID:       cityID,
		OriginCityOther:    ptrOrNil(in.Profile.OriginCityOther),
		NearestAirportID:   airportID,
		DietaryRestriction: ptrOrNil(in.Profile.DietaryRestriction),
		PhoneNumber:        ptrOrNil(in.Profile.PhoneNumber),
		Region:             ptrOrNil(in.Profile.Region),
		Cabang:             ptrOrNil(in.Profile.Cabang),
		Position:           ptrOrNil(in.Profile.Position),
		KtpNumber:          ptrOrNil(in.Profile.KtpNumber),
		NomorKtp:           ptrOrNil(in.Profile.NomorKtp),
		PassportNumber:     ptrOrNil(in.Profile.PassportNumber),
		PassportExpiry:     parseDate(in.Profile.PassportExpiry),
		BlazerSize:         ptrOrNil(in.Profile.BlazerSize),
		NomorMeja:          ptrOrNil(in.Profile.NomorMeja),
		Description:        ptrOrNil(in.Profile.Description),
	}
	user.CreatedBy = in.ActorID

	if err := s.repo.Create(&user); err != nil {
		return nil, ErrDuplicateEntry
	}
	return &user, nil
}

type UpdateInput struct {
	Name     string
	Email    string
	Password string
	Status   string
	// AttendanceStatus is an admin-only manual override of Kehadiran — see
	// the recompute-vs-override branch near the end of Update. Empty means
	// "leave it to the usual auto-recompute", matching every other optional
	// field on this struct.
	AttendanceStatus string
	Profile          ProfileInput
	ActorID          *uint64
}

// Update returns the pre-update snapshot alongside the saved record so
// callers can log a before/after activity diff, matching users.Service.Update.
func (s *Service) Update(uuidStr string, in UpdateInput) (before *users.User, after *users.User, err error) {
	if in.AttendanceStatus != "" && !isValidAttendanceStatus(in.AttendanceStatus) {
		return nil, nil, ErrInvalidAttendanceStatus
	}

	user, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, nil, ErrNotFound
	}
	old := *user
	statusChanged := in.Status != "" && users.UserStatus(in.Status) != old.Status

	if err := s.checkBlazerSizeStock(in.Profile.BlazerSize, old.BlazerSize); err != nil {
		return nil, nil, err
	}

	if in.Name != "" {
		user.Name = in.Name
	}
	if in.Email != "" {
		user.Email = in.Email
	}
	if in.Password != "" {
		// See Create: lowercased before hashing to match PesertaLogin's
		// case-insensitive comparison.
		hashed, err := bcrypt.GenerateFromPassword([]byte(strings.ToLower(in.Password)), 12)
		if err != nil {
			return nil, nil, err
		}
		user.Password = string(hashed)
	}
	if in.Status != "" {
		user.Status = users.UserStatus(in.Status)
	}

	if in.Profile.OriginCityUUID != "" {
		cityID, err := s.repo.ResolveCityID(in.Profile.OriginCityUUID)
		if err != nil {
			return nil, nil, err
		}
		user.OriginCityID = cityID
		// See UpdateProfile for why the stale preloaded association must be
		// cleared too, not just the FK column.
		user.OriginCity = nil
	}
	if in.Profile.NearestAirportUUID != "" {
		airportID, err := s.repo.ResolveAirportID(in.Profile.NearestAirportUUID)
		if err != nil {
			return nil, nil, err
		}
		user.NearestAirportID = airportID
		user.NearestAirport = nil
	}

	user.Title = ptrOrNil(in.Profile.Title)
	user.FirstName = ptrOrNil(in.Profile.FirstName)
	user.MiddleName = ptrOrNil(in.Profile.MiddleName)
	user.LastName = ptrOrNil(in.Profile.LastName)
	user.BirthDate = parseDate(in.Profile.BirthDate)
	user.OriginCityOther = ptrOrNil(in.Profile.OriginCityOther)
	user.DietaryRestriction = ptrOrNil(in.Profile.DietaryRestriction)
	user.PhoneNumber = ptrOrNil(in.Profile.PhoneNumber)
	user.Region = ptrOrNil(in.Profile.Region)
	user.Cabang = ptrOrNil(in.Profile.Cabang)
	user.Position = ptrOrNil(in.Profile.Position)
	user.KtpNumber = ptrOrNil(in.Profile.KtpNumber)
	user.NomorKtp = ptrOrNil(in.Profile.NomorKtp)
	user.PassportNumber = ptrOrNil(in.Profile.PassportNumber)
	user.PassportExpiry = parseDate(in.Profile.PassportExpiry)
	user.BlazerSize = ptrOrNil(in.Profile.BlazerSize)
	user.NomorMeja = ptrOrNil(in.Profile.NomorMeja)
	user.Description = ptrOrNil(in.Profile.Description)
	user.UpdatedBy = in.ActorID
	if in.AttendanceStatus != "" {
		status := in.AttendanceStatus
		user.AttendanceStatus = &status
	} else {
		recomputeAttendanceStatus(user)
	}

	if err := s.repo.Save(user); err != nil {
		return nil, nil, err
	}

	if statusChanged {
		s.invalidateSessions(user.ID)
	}

	return &old, user, nil
}

// UpdateProfile is the self-service counterpart to Update: it only ever
// touches the peserta profile fields (never name/email/password/status),
// since those aren't something the website lets a participant change about
// themselves. Looked up by numeric user ID (from the session) rather than
// UUID (from an admin-supplied path param).
// SelfProfileInput is a true partial-update shape for the self-service
// endpoint: every field is a *string, so "omitted from the JSON body" (nil)
// and "sent as an explicit empty string" (non-nil, pointing to "") are
// distinguishable. The website saves the Form and Shirt Size tabs as two
// separate requests, each only carrying its own fields — unlike admin's
// Update/ProfileInput (always resent as one complete form), an unconditional
// overwrite here would silently null out whichever tab wasn't just saved.
type SelfProfileInput struct {
	Title              *string
	FirstName          *string
	MiddleName         *string
	LastName           *string
	BirthDate          *string
	OriginCityUUID     *string
	OriginCityOther    *string
	NearestAirportUUID *string
	DietaryRestriction *string
	PhoneNumber        *string
	// KtpNumber (the login NIK) is deliberately absent here — it's set at
	// account creation/import and isn't something the participant can change
	// via their own self-service profile form. NomorKtp is the separate,
	// optional KTP-card-number field the website's form actually edits.
	NomorKtp       *string
	PassportNumber *string
	PassportExpiry *string
	BlazerSize     *string
	// PassportSingleName mirrors User.PassportSingleName — see its comment
	// on the model. *bool for the same partial-update reason as every other
	// field here: nil means "not part of this save".
	PassportSingleName *bool
}

func (s *Service) UpdateProfile(userID uint64, in SelfProfileInput) (*users.User, error) {
	user, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, ErrNotFound
	}

	if in.OriginCityUUID != nil {
		cityID, err := s.repo.ResolveCityID(*in.OriginCityUUID)
		if err != nil {
			return nil, err
		}
		user.OriginCityID = cityID
		// user.OriginCity is still the OLD preloaded association from
		// FindByUserID above — GORM re-syncs a belongs-to FK column from its
		// association struct on save, which would silently overwrite the
		// OriginCityID we just set (e.g. clearing it back to the previous
		// city) unless the stale association is cleared too.
		user.OriginCity = nil
	}
	if in.NearestAirportUUID != nil {
		airportID, err := s.repo.ResolveAirportID(*in.NearestAirportUUID)
		if err != nil {
			return nil, err
		}
		user.NearestAirportID = airportID
		user.NearestAirport = nil
	}

	if in.Title != nil {
		user.Title = ptrOrNil(*in.Title)
	}
	if in.FirstName != nil {
		user.FirstName = ptrOrNil(*in.FirstName)
	}
	if in.MiddleName != nil {
		user.MiddleName = ptrOrNil(*in.MiddleName)
	}
	if in.LastName != nil {
		user.LastName = ptrOrNil(*in.LastName)
	}
	if in.PassportSingleName != nil {
		user.PassportSingleName = *in.PassportSingleName
	}
	if in.BirthDate != nil {
		user.BirthDate = parseDate(*in.BirthDate)
	}
	if in.OriginCityOther != nil {
		user.OriginCityOther = ptrOrNil(*in.OriginCityOther)
	}
	if in.DietaryRestriction != nil {
		user.DietaryRestriction = ptrOrNil(*in.DietaryRestriction)
	}
	if in.PhoneNumber != nil {
		user.PhoneNumber = ptrOrNil(*in.PhoneNumber)
	}
	if in.NomorKtp != nil {
		user.NomorKtp = ptrOrNil(*in.NomorKtp)
	}
	if in.PassportNumber != nil {
		user.PassportNumber = ptrOrNil(*in.PassportNumber)
	}
	if in.PassportExpiry != nil {
		expiry := parseDate(*in.PassportExpiry)
		if expiry != nil && expiry.Before(minPassportExpiry) {
			return nil, ErrPassportExpiryTooSoon
		}
		user.PassportExpiry = expiry
	}
	if in.BlazerSize != nil {
		if err := s.checkBlazerSizeStock(*in.BlazerSize, user.BlazerSize); err != nil {
			return nil, err
		}
		user.BlazerSize = ptrOrNil(*in.BlazerSize)
	}
	recomputeAttendanceStatus(user)

	if err := s.repo.Save(user); err != nil {
		return nil, err
	}
	return user, nil
}

// AgreeDataConsent records that the current peserta clicked "Setuju &
// Lanjutkan" on the website's Persetujuan Data Pribadi popup — set once,
// never cleared, so Landing.jsx only shows the popup on a peserta's very
// first login (see users.User.DataConsentAt).
func (s *Service) AgreeDataConsent(userID uint64) (*users.User, error) {
	user, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, ErrNotFound
	}
	now := time.Now()
	user.DataConsentAt = &now
	if err := s.repo.Save(user); err != nil {
		return nil, err
	}
	return user, nil
}

// SetAttendance records the participant's answer to the post-login
// attendance prompt on the website, resolving it straight to the combined
// status (skipping the never-really-observable "hadir but already complete"
// vs "hadir but not complete" ambiguity by just checking completeness now).
func (s *Service) SetAttendance(userID uint64, confirmed bool) (*users.User, error) {
	user, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, ErrNotFound
	}
	status := StatusTidakHadir
	if confirmed {
		status = StatusHadirBelumLengkap
		if isFormComplete(user) && isShirtComplete(user) {
			status = StatusHadirLengkap
		}
	}
	user.AttendanceStatus = &status
	if err := s.repo.Save(user); err != nil {
		return nil, err
	}
	return user, nil
}

// UploadKtpSelf mirrors UploadKtp but resolves the target user from the
// session (numeric ID) rather than an admin-supplied UUID path param.
func (s *Service) UploadKtpSelf(userID uint64, file multipart.File, header *multipart.FileHeader) (string, error) {
	user, err := s.repo.FindByUserID(userID)
	if err != nil {
		return "", ErrNotFound
	}

	path, err := s.storage.Upload(file, header, "ktp")
	if err != nil {
		return "", err
	}

	user.KtpFile = &path
	recomputeAttendanceStatus(user)
	if err := s.repo.Save(user); err != nil {
		return "", err
	}
	return s.storage.GetURL(path), nil
}

// ResetProfile clears every field the website's self-service forms can
// write — everything SelfProfileInput covers (Form + Shirt Size tabs,
// KTP/passport uploads) plus AttendanceStatus, reset back to
// "belum_konfirmasi" — so the peserta looks exactly like they haven't
// logged in and filled anything out yet. Admin-managed fields (Name, Email,
// KtpNumber/NIK, Password, Status, NomorMeja, Description, Region, Cabang,
// Position) are deliberately untouched, matching SelfProfileInput's own
// field set. The old KtpFile/PassportFile paths are only cleared from the
// DB column, not deleted from storage — same as every other upload path in
// this codebase (see storage.StorageInterface), which never deletes a
// replaced file either.
func (s *Service) ResetProfile(uuidStr string, actorID *uint64) (before *users.User, after *users.User, err error) {
	user, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, nil, ErrNotFound
	}
	old := *user

	user.Title = nil
	user.FirstName = nil
	user.MiddleName = nil
	user.LastName = nil
	user.BirthDate = nil
	user.OriginCityID = nil
	user.OriginCity = nil
	user.OriginCityOther = nil
	user.NearestAirportID = nil
	user.NearestAirport = nil
	user.DietaryRestriction = nil
	user.PhoneNumber = nil
	user.NomorKtp = nil
	user.PassportNumber = nil
	user.PassportExpiry = nil
	user.PassportFile = nil
	user.KtpFile = nil
	user.BlazerSize = nil
	belumKonfirmasi := StatusBelumKonfirmasi
	user.AttendanceStatus = &belumKonfirmasi
	user.UpdatedBy = actorID

	if err := s.repo.Save(user); err != nil {
		return nil, nil, err
	}
	return &old, user, nil
}

func (s *Service) Delete(uuidStr string, actorID *uint64) (*users.User, error) {
	user, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, ErrNotFound
	}
	if err := s.repo.Delete(user, actorID); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) UpdateStatus(uuidStr, status string) (oldStatus users.UserStatus, user *users.User, err error) {
	user, err = s.repo.FindByUUID(uuidStr)
	if err != nil {
		return "", nil, ErrNotFound
	}
	oldStatus = user.Status
	user.Status = users.UserStatus(status)
	if err := s.repo.Save(user); err != nil {
		return "", nil, err
	}
	if oldStatus != user.Status {
		s.invalidateSessions(user.ID)
	}
	return oldStatus, user, nil
}

func (s *Service) UploadKtp(uuidStr string, file multipart.File, header *multipart.FileHeader) (string, error) {
	user, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return "", ErrNotFound
	}

	path, err := s.storage.Upload(file, header, "ktp")
	if err != nil {
		return "", err
	}

	user.KtpFile = &path
	recomputeAttendanceStatus(user)
	if err := s.repo.Save(user); err != nil {
		return "", err
	}
	return s.storage.GetURL(path), nil
}

// UploadPassportSelf mirrors UploadPassport but resolves the target user
// from the session (numeric ID) rather than an admin-supplied UUID param.
func (s *Service) UploadPassportSelf(userID uint64, file multipart.File, header *multipart.FileHeader) (string, error) {
	user, err := s.repo.FindByUserID(userID)
	if err != nil {
		return "", ErrNotFound
	}

	path, err := s.storage.Upload(file, header, "passport")
	if err != nil {
		return "", err
	}

	user.PassportFile = &path
	recomputeAttendanceStatus(user)
	if err := s.repo.Save(user); err != nil {
		return "", err
	}
	return s.storage.GetURL(path), nil
}

func (s *Service) UploadPassport(uuidStr string, file multipart.File, header *multipart.FileHeader) (string, error) {
	user, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return "", ErrNotFound
	}

	path, err := s.storage.Upload(file, header, "passport")
	if err != nil {
		return "", err
	}

	user.PassportFile = &path
	recomputeAttendanceStatus(user)
	if err := s.repo.Save(user); err != nil {
		return "", err
	}
	return s.storage.GetURL(path), nil
}
