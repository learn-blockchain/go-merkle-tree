package merkle

import (
	"bytes"
	"testing"
)

type expected struct {
	toPass bool
	hash   []byte
}

type tests struct {
	input    Data
	expected expected
}

func TestMerkleTrees(t *testing.T) {
	i1 := Data{
		[]byte("Hello"),
		[]byte("World"),
	}
	i2 := Data{
		[]byte("Hello"),
		[]byte("World"),
		[]byte("Words"),
		[]byte("Random"),
		[]byte("Things"),
		[]byte("Foos"),
		[]byte("Bars"),
		[]byte("The End"),
	}

	e1 := expected{
		toPass: true,
		hash:   []byte{125, 205, 40, 68, 85, 28, 149, 178, 127, 130, 77, 173, 179, 62, 229, 66, 3, 224, 25, 19, 99, 46, 162, 127, 200, 93, 113, 174, 122, 184, 209, 163},
	}
	e2 := expected{
		toPass: true,
		hash:   []byte{83, 5, 83, 253, 71, 146, 218, 208, 45, 29, 47, 170, 88, 59, 185, 211, 66, 30, 152, 93, 112, 75, 118, 62, 208, 14, 40, 82, 25, 7, 68, 151},
	}

	var tests = []tests{
		// small
		{
			input:    i1,
			expected: e1,
		},
		// larger
		{
			input:    i2,
			expected: e2,
		},
	}

	for idx, tt := range tests {
		n, err := New(tt.input)
		if err != nil {
			t.Errorf("test #%d errored; err: %v", idx+1, err)
		}

		assert := bytes.Equal(n.Hash, tt.expected.hash)
		assert = (assert == tt.expected.toPass)

		if !assert {
			t.Errorf("test #%d failed; input: %v, expected: %v, received: %v", idx+1, tt.input, tt.expected, n)
		}
	}
}
