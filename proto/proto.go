package proto

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"io"
)

func pkcs7Pad(data []byte, blockSize int) []byte {
	pad := blockSize - len(data)%blockSize
	padding := bytes.Repeat([]byte{byte(pad)}, pad)
	return append(data, padding...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}
	pad := data[len(data)-1]
	if int(pad) > len(data) {
		return nil, errors.New("invalid padding")
	}
	return data[:len(data)-int(pad)], nil
}

// ==== CBC 加密 ====

func encryptAES_CBC(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	padded := pkcs7Pad(data, block.BlockSize())
	cipherText := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText, padded)
	return cipherText, nil
}

func decryptAES_CBC(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(data)%block.BlockSize() != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}
	plainText := make([]byte, len(data))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plainText, data)
	return pkcs7Unpad(plainText)
}

// ==== 编码帧（包含随机 IV） ====

func EncodeFrame(opcode uint16, payload any, key []byte) ([]byte, error) {
	// gob 编码
	var plainBuf bytes.Buffer
	if err := gob.NewEncoder(&plainBuf).Encode(payload); err != nil {
		return nil, err
	}

	// gzip 压缩
	compressed, err := gzipCompress(plainBuf.Bytes())
	if err != nil {
		return nil, err
	}

	// 随机 IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// AES-CBC 加密
	encrypted, err := encryptAES_CBC(compressed, key, iv)
	if err != nil {
		return nil, err
	}

	// 构造最终帧：[IV][Opcode][Len][EncryptedData]
	var frame bytes.Buffer
	frame.Write(iv)
	binary.Write(&frame, binary.BigEndian, opcode)
	binary.Write(&frame, binary.BigEndian, uint32(len(encrypted)))
	frame.Write(encrypted)

	return frame.Bytes(), nil
}

// ==== 帧解析（提取 IV, Opcode, EncryptedData） ====

type EncryptedFrame struct {
	IV     []byte
	Opcode uint16
	Data   []byte
}

func ParseFrame(frame []byte) (*EncryptedFrame, error) {
	if len(frame) < 16+2+4 {
		return nil, errors.New("frame too short")
	}
	iv := frame[:16]
	opcode := binary.BigEndian.Uint16(frame[16:18])
	length := binary.BigEndian.Uint32(frame[18:22])
	if len(frame[22:]) < int(length) {
		return nil, errors.New("incomplete data")
	}
	data := frame[22 : 22+length]

	return &EncryptedFrame{
		IV:     iv,
		Opcode: opcode,
		Data:   data,
	}, nil
}

// ==== 解密 payload ====

func DecodePayload(cipherData []byte, iv, key []byte, out any) error {
	// AES 解密
	plainCompressed, err := decryptAES_CBC(cipherData, key, iv)
	if err != nil {
		return err
	}

	// GZIP 解压
	plainData, err := gzipDecompress(plainCompressed)
	if err != nil {
		return err
	}

	// gob 解码
	return gob.NewDecoder(bytes.NewReader(plainData)).Decode(out)
}

func gzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gzipDecompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}
