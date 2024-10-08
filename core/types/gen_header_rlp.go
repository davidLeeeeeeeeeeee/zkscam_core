// Code generated by rlpgen. DO NOT EDIT.

package types

import "github.com/ethereum/go-ethereum/common"
import "github.com/ethereum/go-ethereum/rlp"
import "io"

func (obj *Header) EncodeRLP(_w io.Writer) error {
	w := rlp.NewEncoderBuffer(_w)
	_tmp0 := w.List()
	w.WriteBytes(obj.ParentHash[:])
	w.WriteBytes(obj.UncleHash[:])
	w.WriteBytes(obj.Coinbase[:])
	w.WriteBytes(obj.Root[:])
	w.WriteBytes(obj.TxHash[:])
	w.WriteBytes(obj.ReceiptHash[:])
	w.WriteBytes(obj.Bloom[:])
	if obj.Difficulty == nil {
		w.Write(rlp.EmptyString)
	} else {
		if obj.Difficulty.Sign() == -1 {
			return rlp.ErrNegativeBigInt
		}
		w.WriteBigInt(obj.Difficulty)
	}
	if obj.Number == nil {
		w.Write(rlp.EmptyString)
	} else {
		if obj.Number.Sign() == -1 {
			return rlp.ErrNegativeBigInt
		}
		w.WriteBigInt(obj.Number)
	}
	w.WriteUint64(obj.GasLimit)
	w.WriteUint64(obj.GasUsed)
	w.WriteUint64(obj.Time)
	w.WriteBytes(obj.Extra)
	w.WriteBytes(obj.MixDigest[:])
	w.WriteBytes(obj.Nonce[:])
	_tmp1 := len(obj.MinerAddresses) > 0
	_tmp2 := obj.ZkscamHash != (common.Hash{})
	_tmp3 := len(obj.Signatures) > 0
	_tmp4 := len(obj.BLSPublicKeys) > 0
	_tmp5 := len(obj.AuthBLSSignatures) > 0
	_tmp6 := len(obj.AggregatedSignature) > 0
	_tmp7 := obj.Votes != nil
	_tmp8 := obj.TotalVotes != nil
	_tmp9 := obj.BaseFee != nil
	_tmp10 := obj.WithdrawalsHash != nil
	_tmp11 := obj.BlobGasUsed != nil
	_tmp12 := obj.ExcessBlobGas != nil
	_tmp13 := obj.ParentBeaconRoot != nil
	if _tmp1 || _tmp2 || _tmp3 || _tmp4 || _tmp5 || _tmp6 || _tmp7 || _tmp8 || _tmp9 || _tmp10 || _tmp11 || _tmp12 || _tmp13 {
		_tmp14 := w.List()
		for _, _tmp15 := range obj.MinerAddresses {
			w.WriteBytes(_tmp15[:])
		}
		w.ListEnd(_tmp14)
	}
	if _tmp2 || _tmp3 || _tmp4 || _tmp5 || _tmp6 || _tmp7 || _tmp8 || _tmp9 || _tmp10 || _tmp11 || _tmp12 || _tmp13 {
		w.WriteBytes(obj.ZkscamHash[:])
	}
	if _tmp3 || _tmp4 || _tmp5 || _tmp6 || _tmp7 || _tmp8 || _tmp9 || _tmp10 || _tmp11 || _tmp12 || _tmp13 {
		_tmp16 := w.List()
		for _, _tmp17 := range obj.Signatures {
			w.WriteBytes(_tmp17)
		}
		w.ListEnd(_tmp16)
	}
	if _tmp4 || _tmp5 || _tmp6 || _tmp7 || _tmp8 || _tmp9 || _tmp10 || _tmp11 || _tmp12 || _tmp13 {
		_tmp18 := w.List()
		for _, _tmp19 := range obj.BLSPublicKeys {
			w.WriteBytes(_tmp19)
		}
		w.ListEnd(_tmp18)
	}
	if _tmp5 || _tmp6 || _tmp7 || _tmp8 || _tmp9 || _tmp10 || _tmp11 || _tmp12 || _tmp13 {
		_tmp20 := w.List()
		for _, _tmp21 := range obj.AuthBLSSignatures {
			w.WriteBytes(_tmp21)
		}
		w.ListEnd(_tmp20)
	}
	if _tmp6 || _tmp7 || _tmp8 || _tmp9 || _tmp10 || _tmp11 || _tmp12 || _tmp13 {
		w.WriteBytes(obj.AggregatedSignature)
	}
	if _tmp7 || _tmp8 || _tmp9 || _tmp10 || _tmp11 || _tmp12 || _tmp13 {
		if obj.Votes == nil {
			w.Write(rlp.EmptyString)
		} else {
			if obj.Votes.Sign() == -1 {
				return rlp.ErrNegativeBigInt
			}
			w.WriteBigInt(obj.Votes)
		}
	}
	if _tmp8 || _tmp9 || _tmp10 || _tmp11 || _tmp12 || _tmp13 {
		if obj.TotalVotes == nil {
			w.Write(rlp.EmptyString)
		} else {
			if obj.TotalVotes.Sign() == -1 {
				return rlp.ErrNegativeBigInt
			}
			w.WriteBigInt(obj.TotalVotes)
		}
	}
	if _tmp9 || _tmp10 || _tmp11 || _tmp12 || _tmp13 {
		if obj.BaseFee == nil {
			w.Write(rlp.EmptyString)
		} else {
			if obj.BaseFee.Sign() == -1 {
				return rlp.ErrNegativeBigInt
			}
			w.WriteBigInt(obj.BaseFee)
		}
	}
	if _tmp10 || _tmp11 || _tmp12 || _tmp13 {
		if obj.WithdrawalsHash == nil {
			w.Write([]byte{0x80})
		} else {
			w.WriteBytes(obj.WithdrawalsHash[:])
		}
	}
	if _tmp11 || _tmp12 || _tmp13 {
		if obj.BlobGasUsed == nil {
			w.Write([]byte{0x80})
		} else {
			w.WriteUint64((*obj.BlobGasUsed))
		}
	}
	if _tmp12 || _tmp13 {
		if obj.ExcessBlobGas == nil {
			w.Write([]byte{0x80})
		} else {
			w.WriteUint64((*obj.ExcessBlobGas))
		}
	}
	if _tmp13 {
		if obj.ParentBeaconRoot == nil {
			w.Write([]byte{0x80})
		} else {
			w.WriteBytes(obj.ParentBeaconRoot[:])
		}
	}
	w.ListEnd(_tmp0)
	return w.Flush()
}
