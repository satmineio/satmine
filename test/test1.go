package main

// go run test/test1.go

import (
	"fmt"
	"satmine/satmine"
)

func main() {
	svgData := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100%" height="100%" viewBox="0 0 400 400" mrc721="THIS 721 NAME" mrc721id="11113a884e603a45ae3687285488d6b68dffb5bfe892db16434afbc817b0d22i0">
  <image href="/content/66b871b44ec9519d794df7ef9ce7adaac74584f54a8f745629efc1a1ff7cbe2ei0" x="0" y="0" width="400" height="400" preserveAspectRatio="xMidYMid slice"/>
</svg>`)

	valid, result, err := satmine.ValidateMRC721SvgData(svgData)
	if err != nil {
		fmt.Printf("validateMRC721SvgData错误: %s\n", err)
	} else {
		fmt.Printf("验证结果: %t, 返回信息: %s\n", valid, result)
	}

	name, id, err := satmine.SvgToNameID(svgData)
	if err != nil {
		fmt.Printf("SvgToNameID错误: %s\n", err)
	} else {
		fmt.Printf("MRC721名称: %s, ID: %s\n", name, id)
	}

	imgSrc, err := satmine.SvgToImgSrc(svgData)
	if err != nil {
		fmt.Printf("SvgToImgSrc错误: %s\n", err)
	} else {
		fmt.Printf("Image SRC: %s\n", imgSrc)
	}
}
