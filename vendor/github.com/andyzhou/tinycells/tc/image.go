package tc

import (
	"fmt"
	"github.com/aaparella/carve"
	"github.com/disintegration/gift"
	"golang.org/x/image/draw"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"time"
)

/*
 * Image resize interface
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 //inter macro define
 const (
 	ImgResize = "resize"
 	ImgCropToSize = "cropToSize"
 	ImgSigmoid = "sigmoid"
 )

 //inter default value
 const (
 	ImgDefaultScaleWidth = 240
 	ImgDefaultCropHeight = 120

 	AvatarWidth = 32
 	AvatarHeight = 32
 )

 //image info
 type ImageResize struct {
 	filters []string `gift filters`
 	gfs map[string]*gift.GIFT `gift instances`
 	scaleWidth int `image scale max width`
 	cropHeight int `image crop max height`
 	needCrop bool `image crop or not`
 	tempPath string `image temp path`
 }

 //construct
func NewImageResize(
			scaleWidth, cropHeight int,
			tempPath string,
			needCrop bool,
		) *ImageResize {
	var (
		filters = make(map[string]gift.Filter)
	)

	//check and set default value
	if scaleWidth <= 0 {
		scaleWidth = ImgDefaultScaleWidth
	}
	if cropHeight <= 0 {
		cropHeight = ImgDefaultCropHeight
	}

	//init batch filters
	filters[ImgResize] = gift.Resize(scaleWidth, 0, gift.LanczosResampling)
	if needCrop {
		filters[ImgCropToSize] = gift.CropToSize(scaleWidth, cropHeight, gift.CenterAnchor)
	}

	//self init
	this := &ImageResize{
		filters:make([]string, 0),
		gfs:make(map[string]*gift.GIFT),
		scaleWidth:scaleWidth,
		cropHeight:cropHeight,
		needCrop:needCrop,
		tempPath:tempPath,
	}

	//init batch gift
	for name, filter := range filters {
		this.filters = append(this.filters, name)
		this.gfs[name] = gift.New(filter)
	}

	return this
}

// Decode reads and analyzes the given reader as a GIF image
func (i *ImageResize) SplitAnimatedGif(fileName string, needAll bool) (bool, []string) {
	var (
		frameFile, frameFilePath string
		frameFiles = make([]string, 0)
	)

	if fileName == "" {
		return false, nil
	}

	defer func() {
		if err := recover(); err != nil {
			log.Println("ImageResize::SplitAnimatedGif`rtip、failed, err:", err)
		}
	}()

	//try open file
	filePath := fmt.Sprintf("%s/%s", i.tempPath, fileName)
	file, err := os.Open(filePath)
	if err != nil {
		return false, nil
	}

	//delay close file
	defer file.Close()

	//decode
	gif, err := gif.DecodeAll(file)
	if err != nil {
		log.Println("ImageResize::SplitAnimatedGif failed, err:", err.Error())
		return false, nil
	}

	//frames
	frames := len(gif.Image)

	//get gif width and height
	imgWidth, imgHeight := i.GetGifDimensions(gif)
	overPaintImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(overPaintImage, overPaintImage.Bounds(), gif.Image[0], image.ZP, draw.Src)

	//get now time
	now := time.Now().Unix()

	if !needAll {
		randomFrameIdx := int(time.Now().Unix()) % frames
		frameFile = fmt.Sprintf("%d_%d.png", now, randomFrameIdx)
		frameFilePath = fmt.Sprintf("%s/%s", i.tempPath, frameFile)

		srcImg := gif.Image[randomFrameIdx]
		draw.Draw(overPaintImage, overPaintImage.Bounds(), srcImg, image.ZP, draw.Over)

		//save frame image
		i.SaveImage(frameFilePath, overPaintImage)

		//add into result
		frameFiles = append(frameFiles, frameFile)
		return true, frameFiles
	}

	//process all frames
	for x, srcImg := range gif.Image {
		draw.Draw(overPaintImage, overPaintImage.Bounds(), srcImg, image.ZP, draw.Over)

		// save current frame "stack". This will overwrite an existing file with that name
		frameFile = fmt.Sprintf("%d_%d.png", now, x)
		frameFilePath = fmt.Sprintf("%s/%s", i.tempPath, frameFile)
		i.SaveImage(frameFilePath, overPaintImage)
		//add into result
		frameFiles = append(frameFiles, frameFile)
	}

	return true, frameFiles
}

//get gif dimension
//return x, y width
func (i *ImageResize) GetGifDimensions(gif *gif.GIF) (int, int) {
	var lowestX int
	var lowestY int
	var highestX int
	var highestY int

	for _, img := range gif.Image {
		if img.Rect.Min.X < lowestX {
			lowestX = img.Rect.Min.X
		}
		if img.Rect.Min.Y < lowestY {
			lowestY = img.Rect.Min.Y
		}
		if img.Rect.Max.X > highestX {
			highestX = img.Rect.Max.X
		}
		if img.Rect.Max.Y > highestY {
			highestY = img.Rect.Max.Y
		}
	}

	return highestX - lowestX, highestY - lowestY
}

//resize image from file
func (i *ImageResize) ResizeFromFullFile(fileSrc, fileDst string) bool {
	var (
		bRet bool
	)

	if fileSrc == "" || fileDst == "" {
		return false
	}

	//load src image
	src := i.LoadImage(fileSrc)
	if src == nil {
		return false
	}

	//process image by filters
	for _, name := range i.filters {
		gf, ok := i.gfs[name]
		if !ok {
			continue
		}
		src = i.ProcessOneGift(gf, src)

		//begin save target image
		bRet = i.SaveImage(fileDst, src)
	}

	return bRet
}

//resize image from file
func (i *ImageResize) ResizeFromFile(fileName string) (bool, string) {
	var (
		finalFileName string
		//dst *image.Image
	)

	if fileName == "" {
		return false,  finalFileName
	}

	//format temp file path
	filePath := fmt.Sprintf("%s/%s", i.tempPath, fileName)

	//load original image
	src := i.LoadImage(filePath)
	if src == nil {
		return false, finalFileName
	}

	////reduce height
	//size := src.Bounds().Size()
	//log.Println("size:", size)
	//
	//reSized, err := carve.ReduceHeight(src, 50)
	//if err != nil {
	//	return false, finalFileName
	//}

	//fileName := path.Base(filePath)
	//log.Println("ImageService::ResizeFromFile, tempFile:", tempFile)

	//begin resize and crop image
	for _, name := range i.filters {
		gf, ok := i.gfs[name]
		if !ok {
			continue
		}
		//dst = image.NewNRGBA(gf.Bounds(src.Bounds()))
		//gf.Draw(dst, src)
		src = i.ProcessOneGift(gf, src)
	}

	//dst := image.NewNRGBA(i.gf.Bounds(src.Bounds()))
	//i.gf.Draw(dst, src)

	//generate final file name
	//fileName := path.Base(filePath)
	finalFileName = fmt.Sprintf("%d_%s", time.Now().Unix(), fileName)

	//begin save dst image
	finalTempFile := fmt.Sprintf("%s/%s", i.tempPath, finalFileName)
	bRet := i.SaveImage(finalTempFile, src)
	if !bRet {
		return false, finalFileName
	}

	return true, finalFileName
}

//process one gift
func (i *ImageResize) ProcessOneGift(gf *gift.GIFT, src image.Image) image.Image {
	dst := image.NewNRGBA(gf.Bounds(src.Bounds()))
	gf.Draw(dst, src)
	return dst
}

//reduce height, just testing!!
func (i *ImageResize) ReduceHeight(fileName string) bool {
	//format temp file path
	filePath := fmt.Sprintf("%s/%s", i.tempPath, fileName)

	//load original image
	src := i.LoadImage(filePath)
	if src == nil {
		return false
	}

	//reduce height
	size := src.Bounds().Size()
	//log.Println("size:", size)

	maxHeight := 120
	reducedHeight := int(math.Ceil(float64(size.Y - maxHeight) * 0.5))
	if reducedHeight <= 0 {
		return false
	}

	reSized, err := carve.ReduceHeight(src, reducedHeight)
	if err != nil {
		return false
	}

	//over write?
	bRet := i.SaveImage(filePath, reSized)

	return bRet
}

//resize image from io reader
func (i *ImageResize) ResizeFromIOReader(reader io.Reader) (bool, image.Image) {
	if reader == nil {
		return false, nil
	}

	//create reader
	//reader := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(data))
	//imgSrc, _, err := image.Decode(reader)
	//if err != nil {
	//	log.Println("ImageService::ResizeImageFromByte failed, err:", err.Error())
	//	return false, nil
	//}

	//begin resize image
	//dst := image.NewNRGBA(i.gf.Bounds(imgSrc.Bounds()))
	//i.gf.Draw(dst, imgSrc)

	//size := len(dst.Pix)
	//log.Println("file size:", size)

	//save file
	//dstFilePath := "/Volumes/DATA/project/src/gfs/test.png"
	//bRet := i.saveImage(dstFilePath, dst)

	return true, nil
}

//save dst image file
func (i *ImageResize) SaveImage(filePath string, img image.Image) bool {
	f, err := os.Create(filePath)
	if err != nil {
		log.Println("ImageResize::saveImage, create file ", filePath, " failed, err:", err.Error())
		return false
	}

	//save image data
	defer f.Close()

	tempSlice := strings.Split(filePath, ".")
	tempLen := len(tempSlice)
	kind := tempSlice[tempLen-1]

	//encode image
	switch kind {
	case "jpg":
		{
			err = jpeg.Encode(f, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
			if err != nil {
				log.Println("ImageResize::saveImage, save file ", filePath, " failed, err:", err.Error())
				return false
			}
		}
	default:
		{
			err = png.Encode(f, img)
			if err != nil {
				log.Println("ImageResize::saveImage, save file ", filePath, " failed, err:", err.Error())
				return false
			}
		}
	}

	return true
}

//load original image file
func (i *ImageResize) LoadImage(filePath string) image.Image {
	//try open file
	f, err := os.Open(filePath)
	if err != nil {
		log.Println("ImageResize::loadImage, load file ", filePath, " failed, err:", err.Error())
		return nil
	}
	defer f.Close()

	//try decode image
	img, _, err := image.Decode(f)
	if err != nil {
		log.Println("ImageResize::loadImage, decode file failed, err:", err.Error())
		return nil
	}
	return img
}
