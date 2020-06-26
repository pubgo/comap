package comap

import "testing"

var m = New(1,1)

func BenchmarkGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m.Get(i)
	}
}

func BenchmarkSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m.Set(i, i)
		m.Delete(i)
	}
}

func BenchmarkDel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m.Delete(i)
	}
}
