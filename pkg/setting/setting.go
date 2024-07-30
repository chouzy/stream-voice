package setting

import (
	"fmt"

	"github.com/spf13/viper"
)

type Setting struct {
	vp *viper.Viper
}

func NewSetting(path string) (*Setting, error) {
	vp := viper.New()
	vp.SetConfigFile(path)
	err := vp.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("config file err: %w", err)
	}
	return &Setting{vp: vp}, nil
}

func (s *Setting) ReadSection(k string, v interface{}) error {
	err := s.vp.UnmarshalKey(k, v)
	if err != nil {
		return err
	}
	return nil
}
