package peserta

import (
	"context"
	"errors"
	"mime/multipart"
	"time"

	"baseadmin/backend/modules/users"
	"baseadmin/backend/session"
	"baseadmin/backend/storage"
	"baseadmin/backend/utils"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const dateLayout = "2006-01-02"

var (
	ErrNotFound       = errors.New("peserta not found")
	ErrDuplicateEntry = errors.New("failed to create peserta (email or NIK may already be in use)")
)

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
func isFormComplete(u *users.User) bool {
	dietaryOK := strSet(u.DietaryRestriction) && (*u.DietaryRestriction != "Other" || strSet(u.DietaryRestrictionOther))
	cityOK := u.OriginCityID != nil || strSet(u.OriginCityOther)
	return strSet(u.Title) &&
		strSet(u.FirstName) &&
		strSet(u.LastName) &&
		u.BirthDate != nil &&
		cityOK &&
		u.NearestAirportID != nil &&
		dietaryOK &&
		strSet(u.PhoneNumber) &&
		strSet(u.KtpNumber) &&
		strSet(u.KtpFile) &&
		strSet(u.PassportNumber) &&
		u.PassportExpiry != nil
}

func isShirtComplete(u *users.User) bool {
	return strSet(u.JacketSize) && strSet(u.PoloSize)
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
	repo    *Repository
	storage storage.StorageInterface
	redis   *redis.Client
}

func NewService(repo *Repository, s storage.StorageInterface, rdb *redis.Client) *Service {
	return &Service{repo: repo, storage: s, redis: rdb}
}

// invalidateSessions mirrors users.Service.invalidateSessions — kept for
// parity in case a Peserta account is ever used to log in and its status
// changes.
func (s *Service) invalidateSessions(userID uint64) {
	_ = session.DeleteAllForUser(context.Background(), s.redis, userID)
	_ = s.repo.DeleteAllSessions(userID)
}

func (s *Service) List(p utils.Pagination) ([]UserWithPoints, int64, error) {
	return s.repo.List(p)
}

// ListAll backs CSV/Excel export — every matching row, not one page.
func (s *Service) ListAll(search string) ([]users.User, error) {
	return s.repo.ListAll(search)
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
	Title                   string
	FirstName               string
	MiddleName              string
	LastName                string
	BirthDate               string
	OriginCityUUID          string
	OriginCityOther         string
	NearestAirportUUID      string
	DietaryRestriction      string
	DietaryRestrictionOther string
	PhoneNumber             string
	KtpNumber               string
	NomorKtp                string
	PassportNumber          string
	PassportExpiry          string
	JacketSize              string
	PoloSize                string
	NomorMeja               string
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
	hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), 12)
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
		Name:                    in.Name,
		Email:                   email,
		Password:                string(hashed),
		Status:                  status,
		RoleID:                  &roleID,
		AttendanceStatus:        &attendanceStatus,
		Title:                   ptrOrNil(in.Profile.Title),
		FirstName:               ptrOrNil(in.Profile.FirstName),
		MiddleName:              ptrOrNil(in.Profile.MiddleName),
		LastName:                ptrOrNil(in.Profile.LastName),
		BirthDate:               parseDate(in.Profile.BirthDate),
		OriginCityID:            cityID,
		OriginCityOther:         ptrOrNil(in.Profile.OriginCityOther),
		NearestAirportID:        airportID,
		DietaryRestriction:      ptrOrNil(in.Profile.DietaryRestriction),
		DietaryRestrictionOther: ptrOrNil(in.Profile.DietaryRestrictionOther),
		PhoneNumber:             ptrOrNil(in.Profile.PhoneNumber),
		KtpNumber:               ptrOrNil(in.Profile.KtpNumber),
		NomorKtp:                ptrOrNil(in.Profile.NomorKtp),
		PassportNumber:          ptrOrNil(in.Profile.PassportNumber),
		PassportExpiry:          parseDate(in.Profile.PassportExpiry),
		JacketSize:              ptrOrNil(in.Profile.JacketSize),
		PoloSize:                ptrOrNil(in.Profile.PoloSize),
		NomorMeja:               ptrOrNil(in.Profile.NomorMeja),
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
	Profile  ProfileInput
	ActorID  *uint64
}

// Update returns the pre-update snapshot alongside the saved record so
// callers can log a before/after activity diff, matching users.Service.Update.
func (s *Service) Update(uuidStr string, in UpdateInput) (before *users.User, after *users.User, err error) {
	user, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, nil, ErrNotFound
	}
	old := *user
	statusChanged := in.Status != "" && users.UserStatus(in.Status) != old.Status

	if in.Name != "" {
		user.Name = in.Name
	}
	if in.Email != "" {
		user.Email = in.Email
	}
	if in.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), 12)
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
	user.DietaryRestrictionOther = ptrOrNil(in.Profile.DietaryRestrictionOther)
	user.PhoneNumber = ptrOrNil(in.Profile.PhoneNumber)
	user.KtpNumber = ptrOrNil(in.Profile.KtpNumber)
	user.NomorKtp = ptrOrNil(in.Profile.NomorKtp)
	user.PassportNumber = ptrOrNil(in.Profile.PassportNumber)
	user.PassportExpiry = parseDate(in.Profile.PassportExpiry)
	user.JacketSize = ptrOrNil(in.Profile.JacketSize)
	user.PoloSize = ptrOrNil(in.Profile.PoloSize)
	user.NomorMeja = ptrOrNil(in.Profile.NomorMeja)
	user.UpdatedBy = in.ActorID
	recomputeAttendanceStatus(user)

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
	Title                   *string
	FirstName               *string
	MiddleName              *string
	LastName                *string
	BirthDate               *string
	OriginCityUUID          *string
	OriginCityOther         *string
	NearestAirportUUID      *string
	DietaryRestriction      *string
	DietaryRestrictionOther *string
	PhoneNumber             *string
	// KtpNumber (the login NIK) is deliberately absent here — it's set at
	// account creation/import and isn't something the participant can change
	// via their own self-service profile form. NomorKtp is the separate,
	// optional KTP-card-number field the website's form actually edits.
	NomorKtp       *string
	PassportNumber *string
	PassportExpiry *string
	JacketSize     *string
	PoloSize       *string
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
	if in.BirthDate != nil {
		user.BirthDate = parseDate(*in.BirthDate)
	}
	if in.OriginCityOther != nil {
		user.OriginCityOther = ptrOrNil(*in.OriginCityOther)
	}
	if in.DietaryRestriction != nil {
		user.DietaryRestriction = ptrOrNil(*in.DietaryRestriction)
	}
	if in.DietaryRestrictionOther != nil {
		user.DietaryRestrictionOther = ptrOrNil(*in.DietaryRestrictionOther)
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
		user.PassportExpiry = parseDate(*in.PassportExpiry)
	}
	if in.JacketSize != nil {
		user.JacketSize = ptrOrNil(*in.JacketSize)
	}
	if in.PoloSize != nil {
		user.PoloSize = ptrOrNil(*in.PoloSize)
	}
	recomputeAttendanceStatus(user)

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
