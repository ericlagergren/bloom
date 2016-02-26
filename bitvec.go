package bloom

const (
	_m    = ^0
	_logS = _m>>8&1 + _m>>16&1 + _m>>32&1
	_S    = 1 << _logS

	_W = _S << 3 // word size in bits
)

type bitvec []uint

func (b bitvec) len() int {
	return len(b) * _W
}

func (b bitvec) isSet(pos uint32) bool {
	return b[pos/_W]&(1<<(pos%_W)) != 0
}

func (b bitvec) set(pos uint32) {
	b[pos/_W] |= 1 << (pos % _W)
}

func newBitVec(bits int) bitvec {
	return make(bitvec, bits/_W)
}
