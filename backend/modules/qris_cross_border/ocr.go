package qris_cross_border

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"baseadmin/backend/modules/activity_logs"

	"github.com/anthropics/anthropic-sdk-go"
)

const (
	ocrMaxAttempts = 3
	ocrRetryDelay  = 10 * time.Second

	rejectReasonFailedRead = "Gagal membaca bukti pembayaran setelah 3 kali percobaan"
	rejectReasonDuplicate  = "Duplikat QRIS (No. Referensi sudah terdaftar)"
	rejectReasonTrxFailed  = "Bukti pembayaran menunjukkan transaksi gagal"
	rejectReasonNotQris    = "Gambar bukan bukti pembayaran QRIS"
	rejectReasonNotDanamon = "Bukan bukti pembayaran QRIS Bank Danamon"
	rejectReasonNotMYR     = "Bukan transaksi QRIS Malaysia (tidak ada nominal MYR)"

	acceptedCurrency = "MYR"
)

var errExtractionIncomplete = errors.New("qris extraction did not return a reference number")

// extraction is the shape of extract_qris_data's tool input — the fields
// read back off a QRIS payment-proof screenshot.
type extraction struct {
	IsQrisPayment   bool    `json:"is_qris_payment"`
	IsBankDanamon   bool    `json:"is_bank_danamon"`
	Currency        string  `json:"currency"`
	MerchantName    string  `json:"merchant_name"`
	ReferenceNumber string  `json:"reference_number"`
	NominalAsing    float64 `json:"nominal_asing"`
	NominalIDR      float64 `json:"nominal_idr"`
	TrxStatus       string  `json:"trx_status"`
}

// rejectReason enforces the only proof this program accepts: a Bank Danamon
// QRIS receipt for a Malaysian (MYR) transaction. Checked server-side on top
// of the prompt, so a model that fills a field loosely still can't slip a
// non-MYR or non-Danamon receipt through to approval.
func (e *extraction) rejectReason() string {
	switch {
	case !e.IsQrisPayment:
		return rejectReasonNotQris
	case !e.IsBankDanamon:
		return rejectReasonNotDanamon
	case strings.ToUpper(strings.TrimSpace(e.Currency)) != acceptedCurrency || e.NominalAsing <= 0:
		return rejectReasonNotMYR
	}
	return ""
}

// extractQrisDataTool forces the model's response into this shape via
// tool_choice, so there's no free-text parsing to fall back on.
var extractQrisDataTool = anthropic.ToolUnionParam{
	OfTool: &anthropic.ToolParam{
		Name:        "extract_qris_data",
		Description: anthropic.String("Validate and extract data from a Bank Danamon QRIS cross-border payment proof screenshot for a Malaysian (MYR) transaction."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]any{
				"is_qris_payment": map[string]any{
					"type": "boolean",
					"description": "true HANYA jika gambar adalah bukti/struk pembayaran QRIS (ada logo/tulisan 'QRIS' atau keterangan pembayaran QRIS). " +
						"false jika gambar bukan bukti pembayaran QRIS: transfer bank biasa, virtual account, top-up e-wallet, kartu debit/kredit, " +
						"foto kode QR itu sendiri (bukan struk hasil bayar), foto orang/benda/dokumen, halaman error, layar kosong, atau gambar tidak relevan.",
				},
				"is_bank_danamon": map[string]any{
					"type": "boolean",
					"description": "true HANYA jika bukti pembayaran jelas diterbitkan oleh Bank Danamon (logo/tulisan 'Danamon', 'Bank Danamon', " +
						"atau aplikasi 'D-Bank PRO'). false jika dari bank/aplikasi lain (BCA, Mandiri, BRI, BNI, CIMB, GoPay, OVO, DANA, dll.) " +
						"atau jika asal bank tidak bisa dipastikan dari gambar.",
				},
				"currency": map[string]any{
					"type": "string",
					"description": "Kode mata uang asing (ISO 4217, huruf besar) dari baris 'Nominal', mis. 'MYR', 'THB', 'SGD'. " +
						"Ringgit Malaysia yang ditulis 'RM' diisi 'MYR'. Isi string kosong jika tidak ada nominal dalam mata uang asing " +
						"(mis. hanya ada nominal Rupiah/IDR). Jangan menebak MYR jika tidak tertera di gambar.",
				},
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
					"description": "Nominal dalam mata uang asing sebagai angka polos (tanpa simbol mata uang atau pemisah ribuan), dari baris 'Nominal'. Isi 0 jika tidak ada.",
				},
				"nominal_idr": map[string]any{
					"type":        "number",
					"description": "Nominal konversi ke Rupiah sebagai angka polos, dari baris '(IDR ...)' atau 'Jumlah'.",
				},
				"trx_status": map[string]any{
					"type": "string",
					"enum": []string{"berhasil", "gagal"},
					"description": "Status transaksi sebagaimana ditampilkan PADA screenshot itu sendiri (bukan tebakan dari kelengkapan data). " +
						"Isi 'berhasil' HANYA jika screenshot secara eksplisit menunjukkan indikator sukses: label seperti 'Berhasil', 'Sukses', " +
						"'Transaksi Berhasil', 'Success', 'Completed', atau ikon centang/tanda hijau pada status transaksi. " +
						"Isi 'gagal' jika screenshot menunjukkan indikator gagal/ditolak/dibatalkan/pending/error — label seperti 'Gagal', 'Failed', " +
						"'Ditolak', 'Dibatalkan', 'Cancelled', 'Declined', 'Pending', ikon silang/tanda merah pada status transaksi, ATAU jika gambar " +
						"bukan bukti pembayaran QRIS yang sah (mis. halaman error aplikasi, notifikasi gagal, layar kosong, tangkapan layar tidak " +
						"relevan). Jika tidak ada label status transaksi yang eksplisit sama sekali pada gambar, isi 'gagal' — jangan menganggap " +
						"'berhasil' hanya karena field lain (merchant/nominal/no. referensi) berhasil terbaca.",
				},
			},
			Required: []string{"is_qris_payment", "is_bank_danamon", "currency", "merchant_name", "reference_number", "nominal_asing", "nominal_idr", "trx_status"},
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
				anthropic.NewTextBlock(
					"Periksa gambar ini dan isi tool extract_qris_data. Program ini HANYA menerima bukti pembayaran QRIS lintas negara "+
						"dari Bank Danamon untuk transaksi di Malaysia dengan nominal dalam Ringgit Malaysia (MYR/RM). "+
						"1) Pastikan gambar benar-benar bukti pembayaran QRIS (is_qris_payment). "+
						"2) Pastikan diterbitkan Bank Danamon (is_bank_danamon). "+
						"3) Baca kode mata uang nominal asing apa adanya (currency) — jangan diisi MYR jika yang tertera mata uang lain atau tidak ada nominal asing. "+
						"4) Perhatikan baik-baik status transaksi yang tertulis di screenshot (berhasil/sukses vs gagal/ditolak/dibatalkan/pending/error) — "+
						"jangan asumsikan berhasil hanya karena tampilannya seperti struk resmi; baca label status yang sebenarnya tertera. "+
						"Laporkan apa yang benar-benar terlihat di gambar; jika sebuah field tidak ada, isi string kosong atau 0.",
				),
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
		// A missing reference number on an otherwise-valid receipt is worth
		// retrying (likely a misread); on an image that already fails the
		// QRIS/Danamon/MYR checks it's expected, and retrying would just
		// delay a rejection that's certain anyway.
		if result.ReferenceNumber == "" && result.rejectReason() == "" {
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

	if reason := result.rejectReason(); reason != "" {
		row.MerchantName = nilIfEmpty(result.MerchantName)
		row.ReferenceNumber = nilIfEmpty(result.ReferenceNumber)
		row.NominalAsing = result.NominalAsing
		row.NominalRupiah = result.NominalIDR
		s.finishOCR(row, before, StatusRejected, reason)
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
	trxStatus := strings.ToLower(strings.TrimSpace(result.TrxStatus))
	if trxStatus != "" {
		row.TrxStatus = &trxStatus
	}

	// A screenshot that itself shows a failed/rejected transaction has
	// nothing valid to approve — auto-reject regardless of whether the
	// other fields (merchant/reference/nominal) came back looking complete.
	if trxStatus == TrxStatusGagal {
		s.finishOCR(row, before, StatusRejected, rejectReasonTrxFailed)
		return
	}
	if duplicate {
		s.finishOCR(row, before, StatusRejected, rejectReasonDuplicate)
		return
	}
	// A screenshot that itself shows a successful transaction needs no human
	// review — skip straight to approved instead of waiting_approval.
	if trxStatus == TrxStatusBerhasil {
		s.finishOCR(row, before, StatusApproved, "")
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
