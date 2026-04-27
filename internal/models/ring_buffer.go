package models

// RingBuffer menyimpan N data terakhir untuk keperluan grafik
type RingBuffer struct {
	Data []float64 // Array yang akan menampung history
	size int       // Batas maksimal array (misal: 60)
}

// NewRingBuffer adalah constructor
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		Data: make([]float64, 0, size),
		size: size,
	}
}

// Add memasukkan data baru dan membuang yang terlama jika penuh
func (r *RingBuffer) Add(value float64) {
	if len(r.Data) >= r.size {
		// Buang elemen pertama (index 0) dengan menggeser array
		r.Data = r.Data[1:]
	}
	r.Data = append(r.Data, value)
}