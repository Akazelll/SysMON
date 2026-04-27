# System Resource Monitor (MVP) - Tugas PBO

Aplikasi desktop pemantau sumber daya sistem secara *real-time* (CPU, RAM, Disk, dan Jaringan). Dibangun menggunakan arsitektur berlapis (Layered Architecture) untuk memenuhi syarat Pemrograman Berorientasi Objek.

## 🛠️ Stack Teknologi
* **Bahasa:** Go (Golang) - *Statically-typed & Compiled*
* **GUI Framework:** Fyne v2 (Cross-platform)
* **OS Metrics Library:** gopsutil v3

## 🏗️ Arsitektur Aplikasi (OOP)
Aplikasi ini menerapkan pemisahan layer:
1. **Models:** Struktur data inti (`SystemMetric`, `ProcessStat`, `RingBuffer`).
2. **Repository:** Mengabstraksi pengambilan data kotor dari OS (`os_monitor.go`).
3. **Services:** Logika bisnis yang berdiri sendiri seperti `AlertEngine` dan `Exporter`.
4. **UI (Presentation):** Komponen Fyne kustom seperti `LineChart` dan `ProcessTable`.

## 🚀 Cara Setup dan Menjalankan
1. Pastikan **Go** dan **C Compiler** (GCC/MinGW untuk Windows, atau `build-essential` untuk Linux) sudah terinstal, karena Fyne membutuhkannya untuk merender GUI.
2. Clone repositori ini.
3. Buka terminal di dalam folder proyek, lalu unduh dependensi:
   ```bash
   go mod tidy