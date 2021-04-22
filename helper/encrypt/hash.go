package encrypt

import (
	"github.com/speps/go-hashids"
)

const (
	HashIDSalt   = "l7y4hac4JHaiBkrd52aDASvZSI42vh3J"
	HashIDLength = 15
)

type HashConfig struct {
	Salt   string
	Length int
}

var DefaultHashConfig = HashConfig{
	Salt:   HashIDSalt,
	Length: HashIDLength,
}

type HashConfigOption func(*HashConfig) error

func WithHashSalt(salt string) HashConfigOption {
	return func(cfg *HashConfig) error {
		cfg.Salt = salt
		return nil
	}
}

func WithHashLength(length int) HashConfigOption {
	return func(cfg *HashConfig) error {
		cfg.Length = length
		return nil
	}
}

func newHashIDObject(key, salt string, length int) (*hashids.HashID, error) {
	hd := hashids.NewData()
	hd.Salt = key + salt
	hd.MinLength = length
	return hashids.NewWithData(hd)
}

func applyOptions(cfg *HashConfig, opts []HashConfigOption) error {
	for _, f := range opts {
		if err := f(cfg); err != nil {
			return err
		}
	}
	return nil
}

func EncodeHashID(id int, key string, opts ...HashConfigOption) (string, error) {
	cfg := DefaultHashConfig
	if err := applyOptions(&cfg, opts); err != nil {
		return "", err
	}

	h, err := newHashIDObject(key, cfg.Salt, cfg.Length)
	if err != nil {
		return "", err
	}
	return h.Encode([]int{id})
}

func DecodeHashID(hashID string, key string, opts ...HashConfigOption) (int, error) {
	cfg := DefaultHashConfig
	if err := applyOptions(&cfg, opts); err != nil {
		return 0, err
	}

	h, err := newHashIDObject(key, cfg.Salt, cfg.Length)
	if err != nil {
		return 0, err
	}
	res, err := h.DecodeWithError(hashID)
	if err != nil {
		return 0, err
	}
	return res[0], nil
}
