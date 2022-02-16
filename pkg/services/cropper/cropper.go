package cropper

import (
	"bytes"
	"image/jpeg"

	"github.com/bestleg/ImagePreviewer/pkg/utils"
	"github.com/disintegration/imaging"
	"github.com/pkg/errors"
)

type Transformer interface {
	Crop(img []byte, width, height int, cropFormat uint8) ([]byte, error)
}

type Cropper struct{}

func NewCropper() *Cropper {
	return &Cropper{}
}

func (t *Cropper) Crop(img []byte, width, height int, cropFormat uint8) ([]byte, error) {
	src, err := imaging.Decode(bytes.NewReader(img))
	if err != nil {
		return nil, err
	}
	switch cropFormat {
	case utils.Fill:
		src = imaging.Fill(src, width, height, imaging.Center, imaging.Lanczos)
	case utils.Resize:
		src = imaging.Resize(src, width, height, imaging.Lanczos)
	default:
		return nil, errors.New("failed to take crop format")
	}

	var buff bytes.Buffer
	err = jpeg.Encode(&buff, src, nil)
	return buff.Bytes(), err
}
