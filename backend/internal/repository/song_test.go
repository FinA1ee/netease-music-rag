package repository

import "testing"

func TestFloat32SliceToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []float32
		want string
	}{
		{
			name: "empty",
			in:   []float32{},
			want: "[]",
		},
		{
			name: "multiple values",
			in:   []float32{1, 2.5, -3.25},
			want: "[1.000000,2.500000,-3.250000]",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := float32SliceToString(tc.in)
			if got != tc.want {
				t.Fatalf("float32SliceToString(%v) = %q; want %q", tc.in, got, tc.want)
			}
		})
	}
}
