package helper

import (
	"errors"
	"fmt"
	"github.com/fogleman/gg"
	"image"
	"image/png"
	"os"
)

func BuildImage(fontFile string, backgroundFile string, title string, lineString []string, outfile string) (e error) {
	defer func() {
		if err := recover(); err != nil {
			e = errors.New(fmt.Sprintf("图片创建失败: %#v", err))
		}
	}()
	w := 800
	h := len(lineString)/2*30 + 70

	isDraw := false
	//绘制背景图片
	bgFile, err := os.Open(backgroundFile)
	var _img image.Image = nil
	var (
		x int
		y int
	)
	if err == nil {
		if img, err := png.Decode(bgFile); err == nil {
			//计算右下角高度
			max := img.Bounds().Max
			if max.Y > h {
				h = max.Y + 30
			}
			_img = img
			isDraw = true
			x = w - max.X
			y = h - max.Y
		}
	}

	dc := gg.NewContext(w, h)
	//设置背景颜色
	dc.SetRGB255(88, 180, 185)
	dc.Clear()

	if isDraw {
		//绘制图片
		dc.DrawImage(_img, x, y)
	}

	dc.SetRGB255(255, 255, 255)
	dc.LoadFontFace(fontFile, 34) //加载字体文件
	dc.DrawString(title, 20, 40)
	dc.LoadFontFace(fontFile, 22) //加载字体文件
	for i, item := range lineString {
		l := i / 2
		if i%2 == 0 {
			dc.DrawString(item, 20, 50+float64(l+1)*28)
		} else {
			dc.DrawString(item, 400, 50+float64(l+1)*28)
		}
	}
	return dc.SavePNG(outfile)
}
