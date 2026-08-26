package qris_cross_border

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"baseadmin/backend/modules/activity_logs"

	"github.com/anthropics/anthropic-sdk-go"
)

const (
	ocrMaxAttempts = 3
	ocrRetryDelay  = 10 * time.Second

	rejectReasonFailedRead = "Gagal membaca bukti pembayaran setelah 3 kali percobaan"
	rejectReasonDuplicate  = "Duplikat QRIS (No. Referensi sudah terdaftar)"
)

var errExtractionIncomplete = errors.New("qris extraction did not return a reference number")

// extraction is the shape of extract_qris_data's tool input — the fields
// read back off a QRIS payment-proof screenshot.
type extraction struct {
	MerchantName    string  `json:"merchant_name"`
	ReferenceNumber string  `json:"reference_number"`
	NominalAsing    float64 `json:"nominal_asing"`
	NominalIDR      float64 `json:"nominal_idr"`
}

// extractQrisDataTool forces the model's response into this shape via
// tool_choice, so there's no free-text parsing to fall back on.
var extractQrisDataTool = anthropic.ToolUnionParam{
	OfTool: &anthropic.ToolParam{
		Name:        "extract_qris_data",
		Description: anthropic.String("Extract merchant name, reference number, and both nominal amounts from a QRIS cross-border payment proof screenshot (a bank e-receipt)."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]any{
				"merchant_name": map[string]any{
					"type":        "string",
					"description": "Nama merchant tujuan pembayaran, dari baris 'Merchant Tujuan'.",
				},
				"reference_number": map[string]any{
					"type":        "string",
					"description": "Nomor referensi transaksi, dari baris 'No. Referensi'.",
				},
				"nominal_asing": map[string]any{
					"type":        "number",
					"description": "Nominal dalam mata uang asing sebagai angka polos (tanpa simbol mata uang atau pemisah ribuan), dari baris 'Nominal'.",
				},
				"nominal_idr": map[string]any{
					"type":        "number",
					"description": "Nominal konversi ke Rupiah sebagai angka polos, dari baris '(IDR ...)' atau 'Jumlah'.",
				},
			},
			Required: []string{"merchant_name", "reference_number", "nominal_asing", "nominal_idr"},
		},
	},
}

// extractQrisData sends the proof-of-payment image to the configured OCR
// provider (Claude or DeepSeek — see Service.client) and returns the
// structured extraction.
func (s *Service) extractQrisData(ctx context.Context, imageBytes []byte) (*extraction, error) {
	mediaType := http.DetectContentType(imageBytes)
	encoded := base64.StdEncoding.EncodeToString(imageBytes)

	message, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:      s.model,
		MaxTokens:  1024,
		Thinking:   anthropic.ThinkingConfigParamUnion{OfDisabled: &anthropic.ThinkingConfigDisabledParam{}},
		Tools:      []anthropic.ToolUnionParam{extractQrisDataTool},
		ToolChoice: anthropic.ToolChoiceParamOfTool("extract_qris_data"),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewImageBlockBase64(mediaType, encoded),
				anthropic.NewTextBlock("Ini screenshot bukti pembayaran QRIS lintas negara. Ekstrak data transaksinya lewat tool extract_qris_data."),
			),
		},
	})
	if err != nil {
		return nil, err
	}

	for _, block := range message.Content {
		toolUse, ok := block.AsAny().(anthropic.ToolUseBlock)
		if !ok {
			continue
		}
		var result extraction
		if err := json.Unmarshal(toolUse.Input, &result); err != nil {
			return nil, err
		}
		if result.ReferenceNumber == "" {
			return nil, errExtractionIncomplete
		}
		return &result, nil
	}
	return nil, errExtractionIncomplete
}

// processOCR runs the async extraction pipeline for one just-created record:
// read the proof image, ask Claude to extract the transaction data (retrying
// up to ocrMaxAttempts times, ocrRetryDelay apart), check for a duplicate
// reference number, then write the result back.
//
// Launched as its own goroutine per record (see Service.Create) rather than
// through a shared worker+channel like activity_logs.StartWorker — a dropped
// job here means silently losing a financial reconciliation, which is a
// worse failure mode than the drop-on-full-channel activity_logs accepts for
// log entries, and this admin panel's submission volume doesn't need the
// backpressure a shared worker would add.
func (s *Service) processOCR(uuidStr string) {
	row, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		log.Printf("qris_cross_border: OCR job could not load record %s: %v", uuidStr, err)
		return
	}
	before := toResponse(*row)

	imageBytes, err := s.readImage(row.Image)
	if err != nil {
		log.Printf("qris_cross_border: OCR job could not read image for %s: %v", uuidStr, err)
		s.finishOCR(row, before, StatusRejected, rejectReasonFailedRead)
		return
	}

	var result *extraction
	for attempt := 1; attempt <= ocrMaxAttempts; attempt++ {
		result, err = s.extractQrisData(context.Background(), imageBytes)
		if err == nil {
			break
		}
		log.Printf("qris_cross_border: OCR attempt %d/%d for %s failed: %v", attempt, ocrMaxAttempts, uuidStr, err)
		if attempt < ocrMaxAttempts {
			time.Sleep(ocrRetryDelay)
		}
	}
	if result == nil {
		s.finishOCR(row, before, StatusRejected, rejectReasonFailedRead)
		return
	}

	duplicate, err := s.repo.ReferenceNumberExists(result.ReferenceNumber, row.UUID.String())
	if err != nil {
		log.Printf("qris_cross_border: duplicate check failed for %s: %v", uuidStr, err)
	}

	row.MerchantName = &result.MerchantName
	row.ReferenceNumber = &result.ReferenceNumber
	row.NominalAsing = result.NominalAsing
	row.NominalRupiah = result.NominalIDR

	if duplicate {
		s.finishOCR(row, before, StatusRejected, rejectReasonDuplicate)
		return
	}
	s.finishOCR(row, before, StatusWaitingApproval, "")
}

func (s *Service) readImage(path string) ([]byte, error) {
	reader, err := s.storage.Open(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

// finishOCR persists the job's final status/reason and logs the resulting
// before/after diff. actorID is nil throughout — this is a system process,
// not a user request, so there's no IP/user-agent to attach either.
func (s *Service) finishOCR(row *QrisCrossBorder, before Response, status, reason string) {
	row.Status = status
	row.RejectReason = nilIfEmpty(reason)

	if err := s.repo.Save(row); err != nil {
		log.Printf("qris_cross_border: failed to save OCR result for %s: %v", row.UUID.String(), err)
		return
	}
	after := toResponse(*row)
	activity_logs.LogActivity(nil, activity_logs.ActionUpdate, "qris_cross_border", row.UUID.String(), before, after, "", "")
}
