package settings

import (
	"context"
	"errors"
	"mime/multipart"

	"baseadmin/backend/storage"

	"github.com/redis/go-redis/v9"
)

var ErrInvalidUploadTarget = errors.New("target must be app_logo or app_favicon")

// Service holds business rules for the settings module: key/value upsert and
// invalidating the maintenance-mode cache when it changes.
type Service struct {
	repo    *Repository
	redis   *redis.Client
	storage storage.StorageInterface
}

func NewService(repo *Repository, rdb *redis.Client, s storage.StorageInterface) *Service {
	return &Service{repo: repo, redis: rdb, storage: s}
}

func (s *Service) List() (map[string]string, error) {
	list, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(list))
	for _, item := range list {
		result[item.Key] = item.Value
	}
	return result, nil
}

// publicKeys are readable without a session — pure branding (plus, now, the
// participant-website reference content below), never anything sensitive
// (mail credentials, API keys, etc live under other keys in the same table
// and must stay behind the authenticated List()).
var publicKeys = map[string]bool{
	"app_name":               true,
	"app_logo":               true,
	"app_favicon":            true,
	"primary_color":          true,
	"agenda_file":            true,
	"dress_code_file":        true,
	"event_information_file": true,
	"about_malaysia_file":    true,

	// menu_*_enabled toggles which participant-website home menu tiles are
	// clickable — the site shows a "belum tersedia" popup for any tile whose
	// key is explicitly "false" here (absent/anything else defaults to on).
	"menu_event_agenda_enabled":      true,
	"menu_event_gallery_enabled":     true,
	"menu_dress_code_enabled":        true,
	"menu_qris_cross_border_enabled": true,
	"menu_about_malaysia_enabled":    true,
	"menu_scanner_qr_enabled":        true,
	"menu_history_scanner_enabled":   true,
	"menu_event_information_enabled": true,

	// Deadlines (datetime-local strings, "" = no deadline) the participant
	// website compares against the current time client-side to decide
	// whether to show the attendance prompt / edit-form action at all —
	// backed by a matching server-side check in peserta.Handler so the
	// deadline can't be bypassed by calling the API directly.
	"registration_deadline": true,
	"form_edit_deadline":    true,
}

// Public returns only the whitelisted branding settings, for the login page
// (and anywhere else pre-auth) to render the app name/logo/color correctly
// without needing a session.
func (s *Service) Public() (map[string]string, error) {
	all, err := s.List()
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(publicKeys))
	for key := range publicKeys {
		if v, ok := all[key]; ok {
			result[key] = v
		}
	}
	return result, nil
}

func (s *Service) BulkUpdate(payload map[string]string) error {
	for key, value := range payload {
		if err := s.repo.Upsert(key, value, TypeString); err != nil {
			return err
		}
	}
	if _, ok := payload["maintenance_mode"]; ok {
		s.redis.Del(context.Background(), "settings:maintenance_mode")
	}
	return nil
}

// agenda_file/dress_code_file/event_information_file/about_malaysia_file back
// the participant website's Agenda Acara/Dress Code/Event Information/About
// Malaysia preview popups — an image or PDF, same generic key/value +
// file-upload mechanism as app_logo, just with a different downstream
// consumer.
var uploadTargets = map[string]bool{
	"app_logo":               true,
	"app_favicon":            true,
	"agenda_file":            true,
	"dress_code_file":        true,
	"event_information_file": true,
	"about_malaysia_file":    true,
}

func (s *Service) Upload(target string, file multipart.File, header *multipart.FileHeader) (string, error) {
	if !uploadTargets[target] {
		return "", ErrInvalidUploadTarget
	}

	path, err := s.storage.Upload(file, header, "settings")
	if err != nil {
		return "", err
	}
	if err := s.repo.Upsert(target, path, TypeFile); err != nil {
		return "", err
	}
	return s.storage.GetURL(path), nil
}
