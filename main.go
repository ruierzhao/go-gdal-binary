package main

import (
	"flag"
	"fmt"

	gdal "github.com/lukeroth/gdal"
)

func main() {
	flag.Parse()
	filename := flag.Arg(0)
	if filename == "" {
		fmt.Printf("Usage: tiff [filename]\n")
		return
	}
	buffer := make([]uint8, 256*256)

	driver, err := gdal.GetDriverByName("GTiff")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	dataset := driver.Create(filename, 256, 256, 1, gdal.Byte, nil)
	defer dataset.Close()

	spatialRef := gdal.CreateSpatialReference("")
	spatialRef.FromEPSG(3857)
	srString, err := spatialRef.ToWKT()
	dataset.SetProjection(srString)
	dataset.SetGeoTransform([6]float64{444720, 30, 0, 3751320, 0, -30})
	raster := dataset.RasterBand(1)
	raster.IO(gdal.Write, 0, 0, 256, 256, buffer, 256, 256, 0, 0)
}
