package models

// AlertRule mendefinisikan batasan kapan notifikasi harus muncul
type AlertRule struct {
	Metric    string  // Contoh: "CPU" atau "RAM"
	Threshold float64 // Batas persentase, contoh: 90.0
	Duration  int     // Berapa detik batas tersebut harus terlampaui berturut-turut
}