# DESIGN-WEBADMIN.md

 Semua nilai di bawah ini diambil langsung dari CSS yang dipakai di file HTML utama.

---

## 1. Typography

### Font Family

| Peran | Font | Import |
|---|---|---|
| Body / UI | `DM Sans` | Google Fonts |
| Angka / Kode | `DM Mono` | Google Fonts |

```html
<link href="https://fonts.googleapis.com/css2?family=DM+Sans:ital,opsz,wght@0,9..40,300;0,9..40,400;0,9..40,500;0,9..40,600;1,9..40,400&family=DM+Mono:wght@400;500&display=swap" rel="stylesheet">
```

```css
font-family: 'DM Sans', sans-serif;   /* default body */
font-family: 'DM Mono', monospace;    /* class .mono, angka, kode */
```

### Ukuran & Weight

| Token | Size | Weight | Dipakai untuk |
|---|---|---|---|
| Body default | `14px` | 400 | Teks umum |
| Body small | `13px` | 400 | Isi tabel, form label value |
| Caption | `12px` | 400–500 | Badge, sub-label, pagination |
| Micro | `11px` | 600 | Table header, stat label |
| Tiny | `10px` | 600 | Nav section label |
| Page title | `20px` | 600 | `page-header-title` |
| Modal title | `16px` | 600 | `modal-title` |
| Card title | `14px` | 600 | `card-title` |
| Topbar brand | `17px` | 700 | `topbar-title` |
| Login title | `22px` | 600 | Judul login card |
| Stat value | `24px` | 600 | Angka di stat card, pakai DM Mono |

### Letter Spacing

| Konteks | Nilai |
|---|---|
| Heading / brand | `-.02em` |
| Uppercase label (nav section) | `.08em` |
| Uppercase label (form, stat) | `.05em` – `.07em` |
| Badge & micro label | `.02em` – `.05em` |

### Line Height

```css
line-height: 1.5;   /* default body */
```

---

## 2. Color Palette

### CSS Custom Properties

```css
:root {
  /* Backgrounds */
  --bg:        #f7f6f3;   /* halaman utama */
  --surface:   #ffffff;   /* kartu, sidebar, topbar */
  --surface2:  #f0efe9;   /* hover row, input bg alt */
  --surface3:  #e8e7e0;   /* nav badge bg */

  /* Border */
  --border:    #e2e1db;   /* default border */
  --border2:   #cccbc3;   /* border hover/aktif */

  /* Text */
  --text:      #1a1917;   /* teks utama */
  --text2:     #6b6a65;   /* teks sekunder */
  --text3:     #a09e98;   /* teks placeholder/disabled */

  /* Accent (dark) */
  --accent:    #1a1917;   /* tombol primary, nav aktif, focus border */
  --accent2:   #f0efe9;   /* accent ringan */

  /* Semantic: Success */
  --green:     #2d6a4f;
  --green-bg:  #eaf4ee;

  /* Semantic: Danger */
  --red:       #c0392b;
  --red-bg:    #fdf0ef;

  /* Semantic: Warning */
  --yellow:    #92400e;
  --yellow-bg: #fef3e2;

  /* Semantic: Info */
  --blue:      #1d4ed8;
  --blue-bg:   #eff6ff;
}
```

### Palet Visual

```
Background    ████  #f7f6f3  — warm off-white
Surface       ████  #ffffff  — putih bersih
Surface 2     ████  #f0efe9  — warm gray muda
Surface 3     ████  #e8e7e0  — warm gray sedang
Border        ████  #e2e1db  — outline default
Border 2      ████  #cccbc3  — outline hover

Text          ████  #1a1917  — hampir hitam (warm)
Text 2        ████  #6b6a65  — abu medium
Text 3        ████  #a09e98  — abu muda/placeholder

Accent        ████  #1a1917  — sama dengan --text (dark brand)

Green         ████  #2d6a4f  — teks success
Green BG      ████  #eaf4ee  — bg success
Red           ████  #c0392b  — teks danger
Red BG        ████  #fdf0ef  — bg danger
Yellow        ████  #92400e  — teks warning (amber dark)
Yellow BG     ████  #fef3e2  — bg warning
Blue          ████  #1d4ed8  — teks info
Blue BG       ████  #eff6ff  — bg info
```

---

## 3. Border Radius

```css
--radius:    10px;   /* input, button, badge wrapper */
--radius-lg: 14px;   /* card, stat card */
--radius-xl: 18px;   /* login card, modal */
```

Nilai tambahan yang muncul inline:

| Elemen | Radius |
|---|---|
| Badge | `20px` (pill) |
| Avatar | `50%` (lingkaran) |
| Nav item | `8px` |
| Btn-close | `8px` |
| Form input2 | `9px` |
| Pagination btn | `7px` |
| Btn-icon | `7px` |
| Toast | `10px` |
| Checkbox picker | `5px` |

---

## 4. Shadow

```css
--shadow:    0 1px 3px rgba(0,0,0,.07);    /* card ringan */
--shadow-md: 0 4px 16px rgba(0,0,0,.08);  /* login card, modal */
```

Toast:
```css
box-shadow: 0 4px 20px rgba(0,0,0,.18);
```

---

## 5. Spacing & Layout

### Layout Utama

```css
--sidebar-w: 240px;   /* lebar sidebar */
--nav-h:     60px;    /* tinggi topbar */
```

### Main Content

```css
.main {
  margin-left: 240px;     /* offset sidebar */
  margin-top:  60px;      /* offset topbar */
  padding:     28px;      /* inner padding */
}
```

### Spacing Khas

| Konteks | Nilai |
|---|---|
| Gap antar stat card | `16px` |
| Gap antar elemen page header | `16px` |
| Padding card header | `18px 22px 16px` |
| Padding modal header | `22px 26px 18px` |
| Padding modal body | `24px 26px` |
| Padding modal footer | `16px 26px` |
| Padding table header cell | `11px 20px` |
| Padding table body cell | `13px 20px` |
| Padding filter bar | `14px 22px` |
| Padding login card | `44px 48px` |
| Nav item height | `38px` |

---

## 6. Component Tokens

### Button

```css
/* Primary / Dark */
.btn-dark {
  background: var(--accent);   /* #1a1917 */
  color: #fff;
  height: 38px;
  padding: 0 16px;
  border-radius: var(--radius);   /* 10px */
  font-size: 13px;
  font-weight: 500;
}
.btn-dark:hover { opacity: .88; }

/* Outline */
.btn-outline {
  background: var(--surface);
  color: var(--text2);
  border: 1px solid var(--border);
}
.btn-outline:hover { background: var(--surface2); border-color: var(--border2); }

/* Small modifier */
.btn-sm { height: 32px; padding: 0 12px; font-size: 12px; }

/* Icon only */
.btn-icon { width: 32px; height: 32px; padding: 0; border-radius: 7px; }
```

Tombol primary login (full-width):
```css
.btn-primary {
  width: 100%; height: 48px;
  background: #1a1917; color: #fff;
  border-radius: 10px; font-size: 14px; font-weight: 600;
}
.btn-primary:hover  { opacity: .88; }
.btn-primary:active { transform: scale(.98); }
```

### Badge

```css
.badge {
  display: inline-flex; align-items: center;
  padding: 2px 10px; border-radius: 20px;
  font-size: 12px; font-weight: 500;
}
.badge-green  { background: #eaf4ee; color: #2d6a4f; }
.badge-red    { background: #fdf0ef; color: #c0392b; }
.badge-yellow { background: #fef3e2; color: #92400e; }
.badge-blue   { background: #eff6ff; color: #1d4ed8; }
.badge-gray   { background: #f0efe9; color: #6b6a65; }
```

### Input / Form

```css
.form-input {
  height: 46px;
  border: 1.5px solid var(--border);
  border-radius: 10px;
  font-size: 14px;
}
.form-input:focus { border-color: #1a1917; }

/* Versi compact (modal/form dalam card) */
.form-input2, .form-select2 {
  height: 42px;
  border: 1.5px solid var(--border);
  border-radius: 9px;
  font-size: 13px;
}
.form-textarea2 { height: 80px; resize: vertical; }
```

Label form:
```css
.form-label {
  font-size: 12px; font-weight: 500;
  text-transform: uppercase; letter-spacing: .05em;
  color: var(--text2);
}
```

Search input:
```css
.search-input {
  height: 36px;
  padding: 0 12px 0 34px;   /* kiri untuk ikon search */
  border: 1.5px solid var(--border);
  border-radius: 10px;
  font-size: 13px;
}
```

### Card

```css
.card {
  background: #fff;
  border: 1px solid var(--border);
  border-radius: 14px;
  box-shadow: 0 1px 3px rgba(0,0,0,.07);
  overflow: hidden;
}
```

### Stat Card

```css
.stat-card {
  background: #fff;
  border: 1px solid var(--border);
  border-radius: 14px;
  padding: 20px 22px;
}
.stat-label { font-size: 11px; font-weight: 600; text-transform: uppercase; letter-spacing: .07em; color: #a09e98; }
.stat-value { font-size: 24px; font-weight: 600; font-family: 'DM Mono', monospace; letter-spacing: -.02em; }
.stat-sub   { font-size: 12px; color: #a09e98; margin-top: 5px; }
```

Stat icon:
```css
.stat-icon { width: 36px; height: 36px; border-radius: 9px; }
.stat-icon.green  { background: #eaf4ee; color: #2d6a4f; }
.stat-icon.blue   { background: #eff6ff; color: #1d4ed8; }
.stat-icon.yellow { background: #fef3e2; color: #92400e; }
.stat-icon.red    { background: #fdf0ef; color: #c0392b; }
```

### Modal

```css
.modal-overlay {
  background: rgba(0,0,0,.25);
  backdrop-filter: blur(2px);
}
.modal {
  background: #fff;
  border-radius: 18px;
  max-width: 540px;
  max-height: 90vh;
  animation: modalIn .18s ease;
}
@keyframes modalIn {
  from { opacity: 0; transform: translateY(10px) scale(.98); }
  to   { opacity: 1; transform: none; }
}
```

### Toast

```css
.toast {
  position: fixed; bottom: 28px; right: 28px;
  background: #1a1917; color: #fff;
  border-radius: 10px; padding: 12px 18px;
  font-size: 13px; font-weight: 500;
  box-shadow: 0 4px 20px rgba(0,0,0,.18);
  animation: toastIn .2s ease;
}
@keyframes toastIn {
  from { opacity: 0; transform: translateY(10px); }
  to   { opacity: 1; transform: none; }
}
```

### Tabel

```css
thead th {
  font-size: 11px; font-weight: 600;
  text-transform: uppercase; letter-spacing: .07em;
  color: #a09e98;
  padding: 11px 20px;
  border-bottom: 1px solid var(--border);
}
tbody td {
  padding: 13px 20px;
  font-size: 13px;
  border-bottom: 1px solid var(--border);
}
tbody tr:hover { background: var(--surface2); }
```

### Navigation (Sidebar)

```css
.nav-item {
  height: 38px; padding: 0 10px;
  border-radius: 8px;
  font-size: 13px; color: var(--text2);
}
.nav-item:hover   { background: var(--surface2); color: var(--text); }
.nav-item.active  { background: #1a1917; color: #fff; font-weight: 500; }

.nav-section-label {
  font-size: 10px; font-weight: 600;
  text-transform: uppercase; letter-spacing: .08em;
  color: #a09e98;
}
```

### Avatar

```css
.avatar {
  width: 30px; height: 30px; border-radius: 50%;
  background: #1a1917; color: #fff;
  font-size: 12px; font-weight: 600;
}
```

### Pagination

```css
.page-btn {
  width: 30px; height: 30px; border-radius: 7px;
  border: 1px solid var(--border);
  font-size: 12px; color: var(--text2);
}
.page-btn:hover  { background: var(--surface2); border-color: var(--border2); }
.page-btn.active { background: #1a1917; color: #fff; border-color: #1a1917; }
```

---

## 7. Grid & Responsive

### Grid Utama

```css
/* Stat cards */
.stat-grid { grid-template-columns: repeat(4, 1fr); gap: 16px; }

/* Quick row (2 panel) */
.quick-row { grid-template-columns: 1fr 1fr; gap: 16px; }

/* Form row */
.form-row  { grid-template-columns: 1fr 1fr; gap: 16px; }
```

### Breakpoints

| Breakpoint | Perubahan utama |
|---|---|
| `≤ 768px` | Sidebar jadi overlay slide-in, `.main` margin-left = 0 & padding 16px, stat-grid → 2 kolom, form-row → 1 kolom, modal → bottom sheet |
| `≤ 480px` | stat-grid → 1 kolom, padding main 12px, login card compact |

---

## 8. Motion & Transition

| Elemen | Durasi | Easing |
|---|---|---|
| Button hover (opacity) | `0.15s` | default |
| Button active (scale) | `0.1s` | default |
| Input border focus | `0.15s` | default |
| Nav item bg/color | `0.12s` | default |
| Topbar button | `0.12s` | default |
| Modal entrance | `0.18s` | `ease` |
| Toast entrance | `0.2s` | `ease` |
| Sidebar mobile slide | `0.22s` | `ease` |
| Table row hover | `0.1s` | default |

---

## 9. Ikon

Seluruh ikon menggunakan **inline SVG** dengan style berikut:

```css
/* Nav icon */
width: 16px; height: 16px;
stroke: currentColor; fill: none;
stroke-width: 1.7;
stroke-linecap: round; stroke-linejoin: round;

/* Button icon */
width: 15px; height: 15px;
stroke: currentColor; fill: none;
stroke-width: 2;
stroke-linecap: round; stroke-linejoin: round;

/* Stat icon */
width: 18px; height: 18px;
stroke-width: 1.8;
```

---

## 10. Scrollbar Custom

```css
::-webkit-scrollbar       { width: 4px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: #cccbc3; border-radius: 4px; }
```

---

## 11. Tone & Estetika

- **Warm neutral** — bukan abu-abu dingin; semua warna punya nuansa kekuningan/sepia.
- **Minimal warna** — aksen utama hitam gelap (`#1a1917`), warna semantik hanya untuk status (green/red/yellow/blue).
- **Kepadatan sedang** — bukan compact tapi juga bukan spacious; cocok untuk data-heavy admin.
- **Tipografi tenang** — DM Sans ringan dan modern, DM Mono untuk angka agar mudah di-scan.
- **Border ringan** — menggunakan border tipis `1px` atau `1.5px`, bukan shadow tebal.