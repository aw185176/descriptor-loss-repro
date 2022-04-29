package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/match"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

var (
	fs        = flag.NewFlagSet("reproman", flag.ExitOnError)
	build     = fs.Bool("build", false, "build the image")
	onlybuild = fs.Bool("build-only", false, "build the image and write to layout")
	read      = fs.Bool("read", false, "read the image from the layout")
)

func main() {
	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	var (
		img v1.Image
		p   layout.Path
		err error
	)

	log.Println("build", *build)
	log.Println("build-only", *onlybuild)
	log.Println("read", *read)

	if *build || *onlybuild {
		img, err = buildImg()
		if err != nil {
			log.Fatal("failed to build image", err)
		}
	}

	if *onlybuild || *read {
		p, err = resolvePath()
		if err != nil {
			log.Fatal(err)
		}
	}

	if *onlybuild {
		if err := p.ReplaceImage(img, specialBoyMatcher(), layout.WithAnnotations(map[string]string{"special-boy": "boy1"})); err != nil {
			log.Fatal("failed to write image to layout", err)
		}
		log.Println("wrote image to layout")
		os.Exit(0)
	}

	if *read {
		idx, err := p.ImageIndex()
		if err != nil {
			log.Fatal(err)
		}
		imgs, err := partial.FindImages(idx, specialBoyMatcher())
		if err != nil {
			log.Fatal(err)
		}
		img = imgs[0]
		// img, err = p.I (digest)
		// if err != nil {
		// 	log.Fatal(err)
		// }
	}

	if err := parseAnnotations(img); err != nil {
		log.Fatal(err)
	}
}

func buildImg() (v1.Image, error) {
	data, err := os.ReadFile("testdata/content.txt")
	if err != nil {
		return nil, err
	}

	return addLayer(empty.Image, data)
}

func addLayer(img v1.Image, data []byte) (v1.Image, error) {
	return mutate.Append(img, mutate.Addendum{
		Layer: static.NewLayer(data, types.OCILayer),
		Annotations: map[string]string{
			"com.special.boy": "very-specialest",
		},
	})
}

func resolvePath() (layout.Path, error) {
	path := "dist"
	p, err := layout.FromPath(path)
	if err != nil {
		p, err = layout.Write(path, empty.Index)
		if err != nil {
			return "", fmt.Errorf("failed to create new index at %s: %w", path, err)
		}
	}
	return p, nil
}

func parseAnnotations(img v1.Image) error {
	layers, err := img.Layers()
	if err != nil {
		return err
	}

	for _, l := range layers {
		desc, err := partial.Descriptor(l)
		if err != nil {
			return err
		}
		log.Println("annotations", desc.Annotations)
		log.Println("size", desc.Size)
	}

	return nil
}

func specialBoyMatcher() match.Matcher {
	return match.Annotation("special-boy", "boy1")
}
